package service

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// fakeAnalyzer simulates a slow OCR provider and records peak concurrency.
type fakeAnalyzer struct {
	delay    time.Duration
	inFlight int32
	maxSeen  int32
}

func (f *fakeAnalyzer) Available() bool { return true }

func (f *fakeAnalyzer) AnalyzeDocument(ctx context.Context, data []byte, mime string) (*OCRResult, error) {
	n := atomic.AddInt32(&f.inFlight, 1)
	for {
		m := atomic.LoadInt32(&f.maxSeen)
		if n <= m || atomic.CompareAndSwapInt32(&f.maxSeen, m, n) {
			break
		}
	}
	time.Sleep(f.delay)
	atomic.AddInt32(&f.inFlight, -1)
	// Encode the file's index into doc_type so the test can verify upload order.
	return &OCRResult{DocType: string(data)}, nil
}

func TestProcessDocumentsSyncConcurrentAndOrdered(t *testing.T) {
	fa := &fakeAnalyzer{delay: 50 * time.Millisecond}
	svc := &AIOCRService{analyzer: fa, logger: zap.NewNop().Sugar()}

	const n = 10
	files := make([]SyncFile, n)
	for i := range files {
		files[i] = SyncFile{FileName: strconv.Itoa(i), ContentType: "image/png", Data: []byte(strconv.Itoa(i))}
	}

	start := time.Now()
	res, err := svc.ProcessDocumentsSync(context.Background(), uuid.Nil, files, "default")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("ProcessDocumentsSync: %v", err)
	}

	if len(res.FileResults) != n {
		t.Fatalf("got %d file results, want %d", len(res.FileResults), n)
	}
	for i := range res.FileResults {
		if res.FileResults[i].Filename != strconv.Itoa(i) {
			t.Errorf("order broken at %d: filename=%q", i, res.FileResults[i].Filename)
		}
		if res.FileResults[i].Status != "completed" {
			t.Errorf("file %d status=%q, want completed", i, res.FileResults[i].Status)
		}
		if res.FileResults[i].DocType != strconv.Itoa(i) {
			t.Errorf("file %d doc_type=%q, want %d (order/result mismatch)", i, res.FileResults[i].DocType, i)
		}
	}

	// Sequential would be n*50ms = 500ms; bounded-concurrent (5 workers) ~= 100ms.
	if elapsed > 300*time.Millisecond {
		t.Errorf("ProcessDocumentsSync took %dms — not running concurrently?", elapsed.Milliseconds())
	}
	// Concurrency must stay within the cap (maxConcurrent=5).
	if fa.maxSeen > 5 {
		t.Errorf("peak concurrency %d exceeded the cap of 5", fa.maxSeen)
	}
}

// paspolAnalyzer returns a fixed passport number for every identity file.
type paspolAnalyzer struct{ paspor string }

func (p *paspolAnalyzer) Available() bool { return true }
func (p *paspolAnalyzer) AnalyzeDocument(ctx context.Context, data []byte, mime string) (*OCRResult, error) {
	return &OCRResult{DocType: "paspor", ExtractedData: ExtractedFields{NoPaspor: p.paspor}}, nil
}

type fakePolicy struct{ m *PolicyManifest }

func (f *fakePolicy) ExtractManifest(ctx context.Context, data []byte, text string) (*PolicyManifest, error) {
	return f.m, nil
}

func TestProcessDocumentsSyncEnrichesFromPolicy(t *testing.T) {
	// Stub pdftotext: the policy file's bytes -> policy text; anything else -> "".
	orig := extractPDFText
	extractPDFText = func(ctx context.Context, data []byte) (string, error) {
		if string(data) == "POLISBYTES" {
			return "MANIFEST ... NO POLIS", nil
		}
		return "", nil
	}
	defer func() { extractPDFText = orig }()

	manifest := &PolicyManifest{
		Asuransi: "PT. ASURANSI ASKRIDA SYARIAH", TanggalInput: "2026-06-17",
		Entries: []PolicyEntry{{NoIdentitas: "X2664222", NoPolis: "POL-43",
			TanggalAwal: "2026-07-01", TanggalAkhir: "2026-07-09"}},
	}
	svc := (&AIOCRService{analyzer: &paspolAnalyzer{paspor: "X2664222"}, logger: zap.NewNop().Sugar()}).
		WithPolicy(&fakePolicy{m: manifest})

	files := []SyncFile{
		{FileName: "paspor.jpg", ContentType: "image/jpeg", Data: []byte("img")},
		{FileName: "polis.pdf", ContentType: "application/pdf", Data: []byte("POLISBYTES")},
	}
	res, err := svc.ProcessDocumentsSync(context.Background(), uuid.Nil, files, "default")
	if err != nil {
		t.Fatalf("ProcessDocumentsSync: %v", err)
	}
	if len(res.Data) != 1 {
		t.Fatalf("expected 1 jamaah row (policy makes no row), got %d", len(res.Data))
	}
	row := res.Data[0].(map[string]any)
	if row["no_polis"] != "POL-43" || row["asuransi"] != "ASURANSI ASKRIDA SYARIAH" ||
		row["tanggal_input_polis"] != "2026-06-17" || row["tanggal_awal_polis"] != "2026-07-01" {
		t.Errorf("row not enriched from policy: %+v", row)
	}
	// The policy file is reported, not turned into a jamaah row.
	var sawPolis bool
	for _, fr := range res.FileResults {
		if fr.Filename == "polis.pdf" {
			sawPolis = true
			if fr.Status != "completed" || fr.DocType != "polis" {
				t.Errorf("polis file_result = %+v", fr)
			}
		}
	}
	if !sawPolis {
		t.Errorf("polis.pdf missing from file_results")
	}
}

func TestProcessDocumentsSyncSeedsRowsFromPolicyOnly(t *testing.T) {
	// Uploading ONLY the policy (no identity scans) must still produce a row per
	// jamaah listed in the manifest, with name/passport/birthdate + insurance.
	orig := extractPDFText
	extractPDFText = func(ctx context.Context, data []byte) (string, error) {
		if string(data) == "POLISBYTES" {
			return "MANIFEST ... NO POLIS", nil
		}
		return "", nil
	}
	defer func() { extractPDFText = orig }()

	manifest := &PolicyManifest{
		Asuransi: "PT. ASURANSI ASKRIDA SYARIAH", TanggalInput: "2026-06-17",
		Entries: []PolicyEntry{
			{Nama: "DWINTA DISTIANE", NoIdentitas: "X8557911", TanggalLahir: "1990-01-01",
				NoPolis: "POL-1", TanggalAwal: "2026-07-01", TanggalAkhir: "2026-07-09"},
			{Nama: "HENDRA MAHPUDIN", NoIdentitas: "X8558076", NoPolis: "POL-2",
				TanggalAwal: "2026-07-01", TanggalAkhir: "2026-07-09"},
		},
	}
	svc := (&AIOCRService{analyzer: &fakeAnalyzer{}, logger: zap.NewNop().Sugar()}).
		WithPolicy(&fakePolicy{m: manifest})

	files := []SyncFile{{FileName: "polis.pdf", ContentType: "application/pdf", Data: []byte("POLISBYTES")}}
	res, err := svc.ProcessDocumentsSync(context.Background(), uuid.Nil, files, "default")
	if err != nil {
		t.Fatalf("ProcessDocumentsSync: %v", err)
	}
	if len(res.Data) != 2 {
		t.Fatalf("policy-only should seed 2 jamaah rows, got %d", len(res.Data))
	}
	r0 := res.Data[0].(map[string]any)
	for k, want := range map[string]string{
		"nama": "DWINTA DISTIANE", "no_paspor": "X8557911", "jenis_identitas": "PASPOR",
		"tanggal_lahir": "1990-01-01", "asuransi": "ASURANSI ASKRIDA SYARIAH",
		"no_polis": "POL-1", "tanggal_input_polis": "2026-06-17",
	} {
		if got, _ := r0[k].(string); got != want {
			t.Errorf("seeded row[%q] = %q, want %q", k, got, want)
		}
	}
}
