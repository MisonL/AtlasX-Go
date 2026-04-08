package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLogsEndpointReturnsStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := discoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := os.MkdirAll(paths.LogsRoot, 0o755); err != nil {
		t.Fatalf("mkdir logs root failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(paths.LogsRoot, "atlas.log"), []byte("atlas-data"), 0o644); err != nil {
		t.Fatalf("write log file failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/logs", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"present":true`)) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"file_count":1`)) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestLogsEndpointRejectsWrongMethod(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/logs", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
