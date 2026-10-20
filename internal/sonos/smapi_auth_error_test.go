package sonos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSMAPI_BeginAuthentication_APIClientIdentityError(t *testing.T) {
	for _, tt := range []struct {
		name     string
		body     string
		wantHint bool
	}{
		{"unregistered caller", `{"error":{"code":403,"status":"PERMISSION_DENIED","message":"Method doesn't allow unregistered callers (callers without established identity). Please use API Key or other form of API consumer identity to call this API."}}`, true},
		{"other permission error", `{"error":{"status":"PERMISSION_DENIED","message":"private account detail"}}`, false},
		{"non JSON response", `<html>private account detail</html>`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			c := &SMAPIClient{
				httpClient:  srv.Client(),
				Service:     MusicServiceDescriptor{ID: "284", Auth: MusicServiceAuthAppLink, SecureURI: srv.URL},
				HouseholdID: "test-household", DeviceID: "test-device", TokenStore: newMemTokenStore(),
			}
			_, err := c.BeginAuthentication(context.Background())
			if err == nil || !strings.Contains(err.Error(), "403 Forbidden") {
				t.Fatalf("expected HTTP status in error, got %v", err)
			}
			if got := strings.Contains(err.Error(), "API client identity"); got != tt.wantHint {
				t.Fatalf("API identity hint = %v, want %v: %v", got, tt.wantHint, err)
			}
			if strings.Contains(err.Error(), "private account detail") {
				t.Fatalf("error exposed response body: %v", err)
			}
		})
	}
}
