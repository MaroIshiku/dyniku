package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MaroIshiku/dyniku/internal/models"
	"github.com/MaroIshiku/dyniku/internal/records"
	"golang.org/x/crypto/bcrypt"
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

	handler := newTestHandler(t, configPath, dataDir)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/config", strings.NewReader(`{
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
	addAuthCookie(t, handler, request)

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

func TestConfigAPIAcceptsNumericNetcupCustomerNumber(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	configPath := filepath.Join(dataDir, "config.json")
	err := os.WriteFile(configPath, []byte(`{"settings":[]}`), os.FileMode(0o666))
	if err != nil {
		t.Fatal(err)
	}

	handler := newTestHandler(t, configPath, dataDir)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/config", strings.NewReader(`{
		"settings": [{
			"provider": "netcup",
			"domain": "sub.example.com",
			"api_key": "api-key",
			"password": "api-password",
			"customer_number": 123456,
			"ip_version": "ipv4"
		}]
	}`))
	response := httptest.NewRecorder()
	addAuthCookie(t, handler, request)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
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

	handler := newTestHandler(t, configPath, dataDir)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/config", strings.NewReader(`{
		"settings": [{
			"provider": "netcup",
			"domain": "sub.example.com",
			"api_key": "api-key",
			"password": "api-password"
		}]
	}`))
	response := httptest.NewRecorder()
	addAuthCookie(t, handler, request)

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

func TestProtectedAPIRequiresSetup(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	handler := newHandler(context.Background(), "", testDB{}, testRunner{},
		filepath.Join(dataDir, "config.json"), dataDir, models.BuildInformation{})
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/status", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusLocked {
		t.Fatalf("expected status %d, got %d: %s", http.StatusLocked, response.Code, response.Body.String())
	}
}

func TestFirstRunSetupCreatesAdminAndClosesRegistration(t *testing.T) {
	t.Setenv("ISHIKU_SETUP_SECRET", "test-setup-secret-with-length")
	t.Setenv("ISHIKU_SETUP_SECRET_FILE", filepath.Join(t.TempDir(), "missing-secret"))

	dataDir := t.TempDir()
	handler := newHandler(context.Background(), "", testDB{}, testRunner{},
		filepath.Join(dataDir, "config.json"), dataDir, models.BuildInformation{})

	first := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/setup", strings.NewReader(`{
		"setup_secret": "test-setup-secret-with-length",
		"admin_display_name": "Dyniku Admin",
		"admin_username": "admin",
		"admin_password": "CorrectHorseBatteryStaple",
		"admin_password_confirm": "CorrectHorseBatteryStaple"
	}`))
	first.Header.Set("Content-Type", "application/json")
	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, firstResponse.Code, firstResponse.Body.String())
	}

	authBytes, err := os.ReadFile(filepath.Join(dataDir, authFileName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(authBytes), "CorrectHorseBatteryStaple") {
		t.Fatal("auth file must not contain plaintext password")
	}

	second := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/setup", strings.NewReader(`{
		"setup_secret": "test-setup-secret-with-length",
		"admin_display_name": "Another Admin",
		"admin_username": "other",
		"admin_password": "AnotherCorrectPassword",
		"admin_password_confirm": "AnotherCorrectPassword"
	}`))
	second.Header.Set("Content-Type", "application/json")
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)
	if secondResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, secondResponse.Code, secondResponse.Body.String())
	}
}

func TestAuthStateMigratesLegacyAuthFile(t *testing.T) {
	t.Parallel()

	primaryDataDir := t.TempDir()
	legacyRoot := t.TempDir()
	legacyDataDir := filepath.Join(legacyRoot, "data")
	if err := os.MkdirAll(legacyDataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestAuthFile(t, legacyDataDir)

	service := newAuthService(primaryDataDir, models.BuildInformation{})
	service.path = filepath.Join(primaryDataDir, authFileName)
	service.dataDir = primaryDataDir
	service.timeNow = time.Now

	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(previousWorkingDir); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	})
	if err := os.Chdir(legacyRoot); err != nil {
		t.Fatal(err)
	}

	file, err := service.load()
	if err != nil {
		t.Fatal(err)
	}
	if file.Admin == nil || file.Admin.Username != "admin" {
		t.Fatalf("expected migrated admin auth, got %+v", file.Admin)
	}
	if _, err := os.Stat(filepath.Join(primaryDataDir, authFileName)); err != nil {
		t.Fatalf("expected migrated auth file in primary data dir: %v", err)
	}
}

func newTestHandler(t *testing.T, configPath, dataDir string) http.Handler {
	t.Helper()
	writeTestAuthFile(t, dataDir)
	handler := newHandler(context.Background(), "", testDB{}, testRunner{}, configPath, dataDir,
		models.BuildInformation{Version: "test", Commit: "test-sha", Created: "test-date"})
	return handler
}

func addAuthCookie(t *testing.T, handler http.Handler, request *http.Request) {
	t.Helper()
	login := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/login", strings.NewReader(`{
		"username": "admin",
		"password": "CorrectHorseBatteryStaple"
	}`))
	login.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, login)
	if response.Code != http.StatusOK {
		t.Fatalf("login failed with %d: %s", response.Code, response.Body.String())
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			request.AddCookie(cookie)
			return
		}
	}
	t.Fatal("login response did not include a session cookie")
}

func writeTestAuthFile(t *testing.T, dataDir string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("CorrectHorseBatteryStaple"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dataDir, authFileName), []byte(`{
		"setup_completed": true,
		"created_at": "2026-07-01T00:00:00Z",
		"admin": {
			"display_name": "Dyniku Admin",
			"username": "admin",
			"password_hash": "`+string(hash)+`",
			"created_at": "2026-07-01T00:00:00Z"
		}
	}`), os.FileMode(0o600))
	if err != nil {
		t.Fatal(err)
	}
}
