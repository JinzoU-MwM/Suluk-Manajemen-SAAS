package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GeminiClient struct {
	apiKey  string
	model   string
	httpCli *http.Client
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
	Config   *geminiConfig   `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string       `json:"text,omitempty"`
	InlineData *geminiImage `json:"inlineData,omitempty"`
}

type geminiImage struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiConfig struct {
	Temperature      float64               `json:"temperature,omitempty"`
	MaxOutputTokens  int                   `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string                `json:"responseMimeType,omitempty"`
	ThinkingConfig   *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

// geminiThinkingConfig with ThinkingBudget=0 disables the gemini-2.5-flash
// "thinking" pass so the entire token budget goes to the JSON answer.
type geminiThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

type ExtractedFields struct {
	Nama             string `json:"nama,omitempty"`
	NIK              string `json:"nik,omitempty"`
	NoPaspor         string `json:"no_paspor,omitempty"`
	NamaPaspor       string `json:"nama_paspor,omitempty"`
	TempatLahir      string `json:"tempat_lahir,omitempty"`
	TanggalLahir     string `json:"tanggal_lahir,omitempty"`
	JenisKelamin     string `json:"jenis_kelamin,omitempty"`
	Alamat           string `json:"alamat,omitempty"`
	Provinsi         string `json:"provinsi,omitempty"`
	Kabupaten        string `json:"kabupaten,omitempty"`
	Kecamatan        string `json:"kecamatan,omitempty"`
	Kelurahan        string `json:"kelurahan,omitempty"`
	Agama            string `json:"agama,omitempty"`
	StatusPerkawinan string `json:"status_perkawinan,omitempty"`
	Pekerjaan        string `json:"pekerjaan,omitempty"`
	Kewarganegaraan  string `json:"kewarganegaraan,omitempty"`
	NoTelepon        string `json:"no_telepon,omitempty"`
	NoHP             string `json:"no_hp,omitempty"`
	GolonganDarah    string `json:"golongan_darah,omitempty"`
	Pendidikan       string `json:"pendidikan,omitempty"`
	ProviderVisa     string `json:"provider_visa,omitempty"`
	NoVisa           string `json:"no_visa,omitempty"`
	TanggalVisa      string `json:"tanggal_visa,omitempty"`
	TanggalVisaAkhir string `json:"tanggal_visa_akhir,omitempty"`
	TanggalPaspor    string `json:"tanggal_paspor,omitempty"`
	KotaPaspor       string `json:"kota_paspor,omitempty"`
	TanggalExpired   string `json:"tanggal_expired,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func NewGeminiClient(apiKey string) *GeminiClient {
	if apiKey == "" {
		return nil
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		// gemini-2.0-flash is quota-0 on the current free-tier key; 2.5-flash works.
		model = "gemini-2.5-flash"
	}
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

// Available reports whether a usable Gemini client is configured (nil-safe).
func (c *GeminiClient) Available() bool { return c != nil && c.apiKey != "" }

var systemPrompts = map[string]string{
	"auto": `Anda adalah sistem OCR untuk dokumen perjalanan Umroh/Haji Indonesia.
Anda akan menerima gambar dokumen identitas (KTP, KK, Paspor, atau Visa).
Analisis gambar dan kembalikan JSON dengan:
1. ` + "`doc_type`" + `: salah satu dari "ktp", "kk", "paspor", "visa", "unknown"
2. ` + "`extracted_data`" + `: objek berisi field-field yang diekstrak
3. ` + "`confidence`" + `: skor keyakinan 0-1

Field yang mungkin (isi sesuai dokumen yang terlihat):
- nama: Nama lengkap
- nik: Nomor Induk Kependudukan (16 digit)
- no_paspor: Nomor paspor
- nama_paspor: Nama pada paspor (jika berbeda)
- tempat_lahir: Tempat lahir
- tanggal_lahir: Tanggal lahir (format YYYY-MM-DD)
- jenis_kelamin: Laki-laki / Perempuan
- alamat: Alamat lengkap
- provinsi: Provinsi
- kabupaten: Kabupaten/Kota
- kecamatan: Kecamatan
- kelurahan: Kelurahan/Desa
- agama: Agama
- status_perkawinan: Status perkawinan
- pekerjaan: Pekerjaan
- kewarganegaraan: Kewarganegaraan
- golongan_darah: Golongan darah
- pendidikan: Pendidikan terakhir
- no_telepon: Nomor telepon
- no_hp: Nomor HP
- tanggal_expired: Tanggal berlaku habis (format YYYY-MM-DD)
- tanggal_paspor: Tanggal terbit paspor (format YYYY-MM-DD)
- kota_paspor: Kota penerbit paspor
- provider_visa: Provider/penerbit visa
- no_visa: Nomor visa
- tanggal_visa: Tanggal terbit visa (format YYYY-MM-DD)
- tanggal_visa_akhir: Tanggal akhir visa (format YYYY-MM-DD)

Kembalikan HANYA JSON, tanpa teks lain. Jika tidak yakin suatu field, jangan sertakan field tersebut.`,
}

func (c *GeminiClient) AnalyzeDocument(ctx context.Context, imageData []byte, mimeType string) (*OCRResult, error) {
	if c == nil {
		return nil, fmt.Errorf("gemini client not configured (GEMINI_API_KEY missing)")
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: systemPrompts["auto"]},
					{
						InlineData: &geminiImage{
							MimeType: mimeType,
							Data:     encoded,
						},
					},
				},
			},
		},
		Config: &geminiConfig{
			Temperature:      0.1,
			MaxOutputTokens:  4096,
			ResponseMimeType: "application/json",
			// Disable 2.5-flash "thinking" so the whole budget goes to the JSON
			// answer (otherwise extraction truncates mid-object).
			ThinkingConfig: &geminiThinkingConfig{ThinkingBudget: 0},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.model)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)
	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		var geminiErr struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if err := json.Unmarshal(respBody, &geminiErr); err == nil && geminiErr.Error.Message != "" {
			return nil, fmt.Errorf("gemini api error (status=%s): %s", geminiErr.Error.Status, geminiErr.Error.Message)
		}
		return nil, fmt.Errorf("gemini api returned status %d", resp.StatusCode)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("parse gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini api error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini returned no candidates")
	}

	candidate := geminiResp.Candidates[0]
	if candidate.FinishReason == "SAFETY" {
		return nil, fmt.Errorf("content blocked by safety filter")
	}

	text := ""
	for _, part := range candidate.Content.Parts {
		text += part.Text
	}

	text = cleanJSONString(text)

	var result OCRResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse extracted data: %w", err)
	}

	return &result, nil
}

type OCRResult struct {
	DocType       string          `json:"doc_type"`
	ExtractedData ExtractedFields `json:"extracted_data"`
	Confidence    float64         `json:"confidence"`
}

func cleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	return s
}

func detectMimeType(fileName string) string {
	name := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".webp"):
		return "image/webp"
	case strings.HasSuffix(name, ".pdf"):
		return "application/pdf"
	default:
		return "image/jpeg"
	}
}
