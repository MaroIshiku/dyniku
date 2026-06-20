package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qdm12/ddns-updater/internal/records"
)

type testDB struct{}

func (testDB) SelectAll() []records.Record { return nil }

type testRunner struct{}

func (testRunner) ForceUpdate(context.Context) []error { return nil }

func TestConfigAPIValidatesBeforeWriting(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	configPath := filepath.Join(dataDir, "config.json")
	err := os.WriteFile(configPath, []byte(`{"settings":[]}`), os.FileMode(0o666))
	if err != nil {
		t.Fatal(err)
	}

	handler := newHandler(context.Background(), "", testDB{}, testRunner{}, configPath, dataDir)
	request := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(`{
		"settings": [{
			"provider": "netcup",
			"domain": "sub.example.com",
			"api_key": "api-key",
			"password": "api-password",
			"customer_number": "123456",
			"ip_version": "ipv4"
		}]
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(configBytes), `"provider": "netcup"`) {
		t.Fatalf("expected netcup config to be written, got:\n%s", configBytes)
	}
}

func TestConfigAPIRejectsInvalidProviderConfig(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	configPath := filepath.Join(dataDir, "config.json")
	original := []byte(`{"settings":[]}`)
	err := os.WriteFile(configPath, original, os.FileMode(0o666))
	if err != nil {
		t.Fatal(err)
	}

	handler := newHandler(context.Background(), "", testDB{}, testRunner{}, configPath, dataDir)
	request := httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader(`{
		"settings": [{
			"provider": "netcup",
			"domain": "sub.example.com",
			"api_key": "api-key",
			"password": "api-password"
		}]
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(configBytes) != string(original) {
		t.Fatalf("invalid config should not replace existing config, got:\n%s", configBytes)
	}
}
