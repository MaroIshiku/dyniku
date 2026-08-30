package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MaroIshiku/dyniku/internal/models"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	authFileName           = "auth.json"
	setupSecretDefaultPath = "/run/secrets/ishiku_setup_secret"
	sessionCookieName      = "dyniku_session"
	sessionTTL             = 7 * 24 * time.Hour
	sessionIdleTTL         = 30 * time.Minute
	setupRateLimitWindow   = 15 * time.Minute
	setupMaxFailures       = 5
	loginRateLimitWindow   = 15 * time.Minute
	loginMaxFailures       = 5
	authDirPerm            = os.FileMode(0o700)
	authFilePerm           = os.FileMode(0o600)
	sessionTokenBytes      = 32
	argon2SaltBytes        = 16
	argon2KeyBytes         = 32
	argon2Memory           = 19 * 1024
	argon2Iterations       = 2
	argon2Parallelism      = 1
	minAdminPasswordLength = 12
	maxAuthBodyBytes       = 128 * 1024
)

var (
	errAuthNotFound                 = errors.New("auth file not found")
	errSetupAlreadyDone             = errors.New("setup already completed")
	errSetupSecretMissing           = errors.New("setup secret is not configured")
	errSetupSecretInvalid           = errors.New("setup secret is invalid")
	errInvalidCredentials           = errors.New("invalid username or password")
	errTooManySetupAttempt          = errors.New("too many failed setup attempts, try again later")
	errSetupSecretRequired          = errors.New("setup secret is required")
	errAdminUsernameRequired        = errors.New("admin username is required")
	errAdminDisplayNameRequired     = errors.New("admin display name is required")
	errAdminPasswordTooShort        = errors.New("admin password must be at least 12 characters")
	errAdminPasswordConfirmMismatch = errors.New("admin password confirmation does not match")
	errAdminPasswordMatchesSecret   = errors.New("admin password must not match setup secret")
	errAdminPasswordPlaceholder     = errors.New("admin password must not be a placeholder")
	errAdminPasswordMatchesUsername = errors.New("admin password must not match username")
	errAdminPasswordMatchesApp      = errors.New("admin password must not match app name")
	errPasswordHashInvalidFormat    = errors.New("invalid password hash format")
	errPasswordHashInvalidParams    = errors.New("invalid password hash parameters")
	errPasswordHashUnsupported      = errors.New("unsupported password hash parameters")
	errPasswordHashInvalidSalt      = errors.New("invalid password salt")
	errPasswordHashInvalidKey       = errors.New("invalid password key")
)

type authService struct {
	path      string
	dataDir   string
	buildInfo models.BuildInformation

	mu            sync.Mutex
	sessions      map[string]authSession
	setupFailures map[string]failureBucket
	loginFailures map[string]failureBucket
	timeNow       func() time.Time
}

type authFile struct {
	SetupCompleted bool          `json:"setup_completed"`
	Admin          *adminAccount `json:"admin,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type adminAccount struct {
	DisplayName  string    `json:"display_name"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type authSession struct {
	Username  string
	ExpiresAt time.Time
	LastSeen  time.Time
}

type securityAuditRecord struct {
	Event     string    `json:"event"`
	Actor     string    `json:"actor"`
	Target    string    `json:"target"`
	Result    string    `json:"result"`
	RequestID string    `json:"request_id"`
	Time      time.Time `json:"time"`
}

type failureBucket struct {
	Count      int
	WindowEnds time.Time
}

type publicUser struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	Initials    string `json:"initials"`
}

type authStateResponse struct {
	SetupRequired   bool               `json:"setup_required"`
	SetupConfigured bool               `json:"setup_configured"`
	SetupSecretKey  string             `json:"setup_secret_key,omitempty"`
	Authenticated   bool               `json:"authenticated"`
	User            *publicUser        `json:"user,omitempty"`
	App             appStateInfo       `json:"app"`
	AdminInfo       *adminInfoResponse `json:"admin_info,omitempty"`
	Message         string             `json:"message,omitempty"`
}

type appStateInfo struct {
	AppID    string `json:"app_id"`
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
}

type adminInfoResponse struct {
	AppVersion    string `json:"app_version"`
	BuildDate     string `json:"build_date"`
	GitHubSHA     string `json:"github_sha"`
	DataDirectory string `json:"data_directory"`
	ConfigPath    string `json:"config_path"`
	Database      string `json:"database_status"`
	SetupState    string `json:"setup_state"`
	Health        string `json:"health_status"`
	LogLevel      string `json:"log_level"`
}

type setupRequest struct {
	SetupSecret          string `json:"setup_secret"`
	AdminDisplayName     string `json:"admin_display_name"`
	AdminUsername        string `json:"admin_username"`
	AdminEmail           string `json:"admin_email"`
	AdminPassword        string `json:"admin_password"`
	AdminPasswordConfirm string `json:"admin_password_confirm"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func newAuthService(dataDir string, buildInfo models.BuildInformation) *authService {
	return &authService{
		path:          filepath.Join(dataDir, authFileName),
		dataDir:       dataDir,
		buildInfo:     buildInfo,
		sessions:      make(map[string]authSession),
		setupFailures: make(map[string]failureBucket),
		loginFailures: make(map[string]failureBucket),
		timeNow:       time.Now,
	}
}

func (h *handlers) authState(w http.ResponseWriter, r *http.Request) {
	state, err := h.makeAuthState(r)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *handlers) setup(w http.ResponseWriter, r *http.Request) {
	if !h.auth.allowSetupAttempt(clientIP(r)) {
		writeSecurityAudit(r, "first_run_setup", "anonymous", "administrator", "rate_limited")
		httpError(w, http.StatusTooManyRequests, errTooManySetupAttempt.Error())
		return
	}

	var request setupRequest
	if err := decodeJSONBody(r, &request); err != nil {
		h.auth.recordSetupFailure(clientIP(r))
		writeSecurityAudit(r, "first_run_setup", "anonymous", "administrator", "invalid_request")
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.auth.createFirstAdmin(request)
	if err != nil {
		h.auth.recordSetupFailure(clientIP(r))
		writeSecurityAudit(r, "first_run_setup", "anonymous", strings.TrimSpace(request.AdminUsername), "failed")
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, errSetupAlreadyDone):
			status = http.StatusConflict
		case errors.Is(err, errSetupSecretMissing):
			status = http.StatusServiceUnavailable
		case errors.Is(err, errSetupSecretInvalid):
			status = http.StatusUnauthorized
		}
		httpError(w, status, err.Error())
		return
	}

	h.auth.clearSetupFailures(clientIP(r))
	writeSecurityAudit(r, "first_run_setup", user.Username, user.Username, "succeeded")
	h.setSessionCookie(w, r, user.Username)
	writeJSON(w, http.StatusCreated, map[string]any{
		"authenticated": true,
		"user":          makePublicUser(user),
	})
}

func (h *handlers) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeSecurityAudit(r, "sign_in", "anonymous", "unknown", "invalid_request")
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}

	loginKey := loginRateLimitKey(request.Username, clientIP(r))
	if !h.auth.allowLoginAttempt(loginKey) {
		writeSecurityAudit(r, "sign_in", "anonymous", strings.TrimSpace(request.Username), "rate_limited")
		httpError(w, http.StatusTooManyRequests, errInvalidCredentials.Error())
		return
	}

	user, err := h.auth.authenticate(request.Username, request.Password)
	if err != nil {
		h.auth.recordLoginFailure(loginKey)
		writeSecurityAudit(r, "sign_in", "anonymous", strings.TrimSpace(request.Username), "failed")
		httpError(w, http.StatusUnauthorized, errInvalidCredentials.Error())
		return
	}

	h.auth.clearLoginFailures(loginKey)
	writeSecurityAudit(r, "sign_in", user.Username, user.Username, "succeeded")
	h.setSessionCookie(w, r, user.Username)
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"user":          makePublicUser(user),
	})
}

func (h *handlers) logout(w http.ResponseWriter, r *http.Request) {
	username, _ := h.auth.usernameFromRequest(r)
	h.auth.destroySessionFromRequest(r)
	writeSecurityAudit(r, "session_revocation", username, username, "succeeded")
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"logged_out": true})
}

func (h *handlers) me(w http.ResponseWriter, r *http.Request) {
	state, err := h.makeAuthState(r)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *handlers) adminInfo(w http.ResponseWriter, r *http.Request) {
	state, err := h.makeAuthState(r)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, state.AdminInfo)
}

func (h *handlers) makeAuthState(r *http.Request) (authStateResponse, error) {
	authFile, err := h.auth.load()
	if err != nil && !errors.Is(err, errAuthNotFound) {
		return authStateResponse{}, err
	}

	_, setupConfigured, setupSecretKey, secretErr := readSetupSecret()
	if secretErr != nil {
		setupConfigured = false
	}

	user, authenticated := h.currentUser(r)
	setupRequired := authFile.Admin == nil || !authFile.SetupCompleted
	response := authStateResponse{
		SetupRequired:   setupRequired,
		SetupConfigured: setupConfigured,
		SetupSecretKey:  setupSecretKey,
		Authenticated:   authenticated,
		User:            user,
		App: appStateInfo{
			AppID:    "dyniku",
			Name:     "Dyniku",
			Subtitle: "DDNS-Updater Web GUI",
		},
	}
	if setupRequired && !setupConfigured {
		response.Message = "Setup secret is missing. Set ISHIKU_SETUP_SECRET " +
			"or use ISHIKU_SETUP_SECRET_FILE in a custom deployment."
	}
	if authenticated {
		response.AdminInfo = h.makeAdminInfo(authFile)
	}
	return response, nil
}

func (h *handlers) makeAdminInfo(authFile authFile) *adminInfoResponse {
	setupState := "required"
	if authFile.SetupCompleted && authFile.Admin != nil {
		setupState = "completed"
	}
	logLevel := os.Getenv("ISHIKU_LOG_LEVEL")
	if logLevel == "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}
	if logLevel == "" {
		logLevel = "info"
	}
	return &adminInfoResponse{
		AppVersion:    h.buildInfo.Version,
		BuildDate:     h.buildInfo.Created,
		GitHubSHA:     h.buildInfo.Commit,
		DataDirectory: h.dataDir,
		ConfigPath:    h.configPath,
		Database:      "ready",
		SetupState:    setupState,
		Health:        "ready",
		LogLevel:      logLevel,
	}
}

func (h *handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authFile, err := h.auth.load()
		if errors.Is(err, errAuthNotFound) || authFile.Admin == nil || !authFile.SetupCompleted {
			httpError(w, http.StatusLocked, "first-run setup is required")
			return
		}
		if err != nil {
			httpError(w, http.StatusInternalServerError, "auth state unavailable")
			return
		}
		if _, ok := h.currentUser(r); !ok {
			httpError(w, http.StatusUnauthorized, "login required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handlers) currentUser(r *http.Request) (*publicUser, bool) {
	username, ok := h.auth.usernameFromRequest(r)
	if !ok {
		return nil, false
	}
	authFile, err := h.auth.load()
	if err != nil || authFile.Admin == nil || authFile.Admin.Username != username {
		return nil, false
	}
	return makePublicUser(*authFile.Admin), true
}

func (h *handlers) setSessionCookie(w http.ResponseWriter, r *http.Request, username string) {
	token := h.auth.createSession(username)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		Expires:  h.auth.timeNow().Add(sessionTTL),
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func (a *authService) load() (authFile, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.loadLocked()
}

func (a *authService) loadLocked() (authFile, error) {
	data, err := os.ReadFile(a.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return authFile{}, errAuthNotFound
		}
		return authFile{}, fmt.Errorf("reading auth state: %w", err)
	}

	var file authFile
	if err := json.Unmarshal(data, &file); err != nil {
		return authFile{}, fmt.Errorf("parsing auth state: %w", err)
	}
	return file, nil
}

func (a *authService) createFirstAdmin(request setupRequest) (adminAccount, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	existing, err := a.loadLocked()
	if err == nil && existing.Admin != nil && existing.SetupCompleted {
		return adminAccount{}, errSetupAlreadyDone
	}
	if err != nil && !errors.Is(err, errAuthNotFound) {
		return adminAccount{}, err
	}

	setupSecret, configured, _, err := readSetupSecret()
	if err != nil {
		return adminAccount{}, err
	}
	if !configured {
		return adminAccount{}, errSetupSecretMissing
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(request.SetupSecret)), []byte(setupSecret)) != 1 {
		return adminAccount{}, errSetupSecretInvalid
	}
	if err := validateAdminPassword(request, setupSecret); err != nil {
		return adminAccount{}, err
	}

	username := strings.TrimSpace(request.AdminUsername)
	displayName := strings.TrimSpace(request.AdminDisplayName)
	hash, err := hashPassword(request.AdminPassword)
	if err != nil {
		return adminAccount{}, fmt.Errorf("hashing admin password: %w", err)
	}

	admin := adminAccount{
		DisplayName:  displayName,
		Username:     username,
		Email:        strings.TrimSpace(request.AdminEmail),
		PasswordHash: hash,
		CreatedAt:    a.timeNow().UTC(),
	}
	file := authFile{
		SetupCompleted: true,
		Admin:          &admin,
		CreatedAt:      a.timeNow().UTC(),
	}
	if err := a.saveLocked(file); err != nil {
		return adminAccount{}, err
	}
	return admin, nil
}

func (a *authService) saveLocked(file authFile) error {
	if err := os.MkdirAll(a.dataDir, authDirPerm); err != nil {
		return fmt.Errorf("creating auth directory: %w", err)
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding auth state: %w", err)
	}
	tempPath := a.path + ".tmp"
	if err := os.WriteFile(tempPath, append(data, '\n'), authFilePerm); err != nil {
		return fmt.Errorf("writing auth state: %w", err)
	}
	if err := os.Rename(tempPath, a.path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("storing auth state: %w", err)
	}
	return nil
}

func (a *authService) authenticate(username, password string) (adminAccount, error) {
	file, err := a.load()
	if err != nil || file.Admin == nil || !file.SetupCompleted {
		return adminAccount{}, errInvalidCredentials
	}
	if file.Admin.Username != strings.TrimSpace(username) {
		return adminAccount{}, errInvalidCredentials
	}
	valid, upgrade, err := verifyPassword(file.Admin.PasswordHash, password)
	if err != nil || !valid {
		return adminAccount{}, errInvalidCredentials
	}
	if upgrade {
		if err := a.upgradePasswordHash(file.Admin.Username, password); err != nil {
			return adminAccount{}, errInvalidCredentials
		}
	}
	return *file.Admin, nil
}

func (a *authService) upgradePasswordHash(username, password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	file, err := a.loadLocked()
	if err != nil || file.Admin == nil || file.Admin.Username != username {
		return errInvalidCredentials
	}
	file.Admin.PasswordHash = hash
	return a.saveLocked(file)
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("creating password salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyBytes)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func verifyPassword(encodedHash, password string) (valid, upgrade bool, err error) {
	if strings.HasPrefix(encodedHash, "$2") {
		if bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password)) != nil {
			return false, false, errInvalidCredentials
		}
		return true, true, nil
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, false, errPasswordHashInvalidFormat
	}
	var version int
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, false, fmt.Errorf("parsing password hash: %w", err)
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, false, fmt.Errorf("parsing password parameters: %w", err)
	}
	if parts[2] != fmt.Sprintf("v=%d", version) ||
		parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, iterations, parallelism) {
		return false, false, errPasswordHashInvalidParams
	}
	if version != argon2.Version || memory < argon2Memory || iterations < argon2Iterations ||
		parallelism < argon2Parallelism || memory > 256*1024 || iterations > 10 || parallelism > 8 {
		return false, false, errPasswordHashUnsupported
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return false, false, errPasswordHashInvalidSalt
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) < 16 || len(expected) > 64 {
		return false, false, errPasswordHashInvalidKey
	}
	keyLength := uint32(len(expected)) //nolint:gosec // Length is validated as 16..64 directly above.
	actual := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(actual, expected) == 1, false, nil
}

func (a *authService) createSession(username string) string {
	tokenBytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		panic(err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.timeNow()
	a.sessions[token] = authSession{Username: username, ExpiresAt: now.Add(sessionTTL), LastSeen: now}
	return token
}

func (a *authService) usernameFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	session, ok := a.sessions[cookie.Value]
	now := a.timeNow()
	if !ok || now.After(session.ExpiresAt) || now.Sub(session.LastSeen) > sessionIdleTTL {
		delete(a.sessions, cookie.Value)
		return "", false
	}
	session.LastSeen = now
	a.sessions[cookie.Value] = session
	return session.Username, true
}

func (a *authService) destroySessionFromRequest(r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, cookie.Value)
}

func (a *authService) allowSetupAttempt(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	bucket := a.setupFailures[ip]
	now := a.timeNow()
	if now.After(bucket.WindowEnds) {
		return true
	}
	return bucket.Count < setupMaxFailures
}

func (a *authService) recordSetupFailure(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.timeNow()
	bucket := a.setupFailures[ip]
	if now.After(bucket.WindowEnds) {
		bucket = failureBucket{WindowEnds: now.Add(setupRateLimitWindow)}
	}
	bucket.Count++
	a.setupFailures[ip] = bucket
}

func (a *authService) clearSetupFailures(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.setupFailures, ip)
}

func (a *authService) allowLoginAttempt(key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	bucket := a.loginFailures[key]
	now := a.timeNow()
	return now.After(bucket.WindowEnds) || bucket.Count < loginMaxFailures
}

func (a *authService) recordLoginFailure(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.timeNow()
	bucket := a.loginFailures[key]
	if now.After(bucket.WindowEnds) {
		bucket = failureBucket{WindowEnds: now.Add(loginRateLimitWindow)}
	}
	bucket.Count++
	a.loginFailures[key] = bucket
}

func (a *authService) clearLoginFailures(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.loginFailures, key)
}

func loginRateLimitKey(username, ip string) string {
	return strings.ToLower(strings.TrimSpace(username)) + "\x00" + ip
}

func writeSecurityAudit(r *http.Request, event, actor, target, result string) {
	if actor == "" {
		actor = "anonymous"
	}
	if target == "" {
		target = "unknown"
	}
	record := securityAuditRecord{
		Event:     event,
		Actor:     actor,
		Target:    target,
		Result:    result,
		RequestID: chimiddleware.GetReqID(r.Context()),
		Time:      time.Now().UTC(),
	}
	data, err := json.Marshal(record)
	if err == nil {
		log.Printf("security_audit %s", data)
	}
}

func readSetupSecret() (secret string, configured bool, source string, err error) {
	source = strings.TrimSpace(os.Getenv("ISHIKU_SETUP_SECRET_FILE"))
	if source == "" {
		source = setupSecretDefaultPath
	}
	//nolint:gosec // Operators intentionally control this Docker secret path.
	if data, readErr := os.ReadFile(source); readErr == nil {
		secret = strings.TrimSpace(string(data))
		return secret, secret != "", "ISHIKU_SETUP_SECRET_FILE", nil
	} else if !errors.Is(readErr, os.ErrNotExist) && os.Getenv("ISHIKU_SETUP_SECRET_FILE") != "" {
		return "", false, "ISHIKU_SETUP_SECRET_FILE", fmt.Errorf("reading setup secret file: %w", readErr)
	}

	secret = strings.TrimSpace(os.Getenv("ISHIKU_SETUP_SECRET"))
	return secret, secret != "", "ISHIKU_SETUP_SECRET", nil
}

func validateAdminPassword(request setupRequest, setupSecret string) error {
	username := strings.TrimSpace(request.AdminUsername)
	displayName := strings.TrimSpace(request.AdminDisplayName)
	password := request.AdminPassword
	normalizedPassword := strings.ToLower(strings.TrimSpace(password))

	switch {
	case strings.TrimSpace(request.SetupSecret) == "":
		return errSetupSecretRequired
	case username == "":
		return errAdminUsernameRequired
	case displayName == "":
		return errAdminDisplayNameRequired
	case len(password) < minAdminPasswordLength:
		return errAdminPasswordTooShort
	case password != request.AdminPasswordConfirm:
		return errAdminPasswordConfirmMismatch
	case password == setupSecret:
		return errAdminPasswordMatchesSecret
	case isPlaceholderPassword(normalizedPassword):
		return errAdminPasswordPlaceholder
	case normalizedPassword == strings.ToLower(username):
		return errAdminPasswordMatchesUsername
	case normalizedPassword == "dyniku":
		return errAdminPasswordMatchesApp
	}
	return nil
}

func isPlaceholderPassword(password string) bool {
	switch password {
	case "changeme", "admin", "password", "passwort", "123456", "ishiku", "change-me":
		return true
	default:
		return false
	}
}

func makePublicUser(user adminAccount) *publicUser {
	return &publicUser{
		DisplayName: user.DisplayName,
		Username:    user.Username,
		Email:       user.Email,
		Initials:    initials(user.DisplayName, user.Username),
	}
}

func initials(displayName, username string) string {
	source := strings.TrimSpace(displayName)
	if source == "" {
		source = strings.TrimSpace(username)
	}
	parts := strings.Fields(source)
	if len(parts) == 0 {
		return "D"
	}
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}
	first := []rune(parts[0])
	second := []rune(parts[1])
	return strings.ToUpper(string([]rune{first[0], second[0]}))
}

func decodeJSONBody(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxAuthBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
