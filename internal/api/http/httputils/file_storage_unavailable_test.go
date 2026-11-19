package httputils_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mini-maxit/backend/internal/api/http/httputils"
	"github.com/mini-maxit/file-storage/pkg/filestorage"
)

func TestHandleServiceError_FileStorageUnavailable(t *testing.T) {
	err := fmt.Errorf("failed to generate signed URL: %w", &filestorage.ErrClient{
		Message: "failed to call /sign endpoint",
		Err:     &net.DNSError{Err: "no such host", Name: "file-storage"},
	})
	w := httptest.NewRecorder()
	httputils.HandleServiceError(w, err, nil, nil)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", w.Code)
	}
}
