package service

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

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
	res, err := svc.ProcessDocumentsSync(context.Background(), files, "default")
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
