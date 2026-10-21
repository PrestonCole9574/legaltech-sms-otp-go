package infrai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallRetriesRateLimitErrorEnvelope(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"ok":false,"error":{"code":"RATE_LIMITED"}}`))
			return
		}
		w.Write([]byte(`{"ok":true,"data":{"verified":true}}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, Key: "test", HTTP: server.Client()}
	verified, err := client.VerifyOTP(context.Background(), "+15550001111", "123456")
	if err != nil || !verified || attempts != 2 {
		t.Fatalf("verified=%v err=%v attempts=%d", verified, err, attempts)
	}
}
