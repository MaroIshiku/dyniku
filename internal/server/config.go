package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	jsonparams "github.com/qdm12/ddns-updater/internal/params"
)

const maxConfigBytes = 1024 * 1024

type configResponse struct {
	Path        string          `json:"path"`
	EnvConfig   bool            `json:"env_config"`
	Config      json.RawMessage `json:"config"`
	RestartHint string          `json:"restart_hint"`
	Warnings    []string        `json:"warnings,omitempty"`
}

func (h *handlers) getConfig(w http.ResponseWriter, _ *http.Request) {
	configBytes, err := os.ReadFile(h.configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configBytes = []byte(`{"settings":[]}`)
		} else {
			httpError(w, http.StatusInternalServerError, "reading config: "+err.Error())
			return
		}
	}

	normalized, err := normalizeJSON(configBytes)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "config is not valid JSON: "+err.Error())
		return
	}
	warnings, validationErr := jsonparams.ValidateJSON(normalized)

	writeJSON(w, http.StatusOK, configResponse{
		Path:        h.configPath,
		EnvConfig:   os.Getenv("CONFIG") != "",
		Config:      normalized,
		RestartHint: "Restart the container after saving so the updater reloads data/config.json.",
		Warnings:    appendValidationWarning(warnings, validationErr),
	})
}

func (h *handlers) putConfig(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("CONFIG") != "" {
		httpError(w, http.StatusConflict,
			"CONFIG environment variable is set; file edits would be overwritten on restart")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxConfigBytes+1))
	if err != nil {
		httpError(w, http.StatusBadRequest, "reading request body: "+err.Error())
		return
	}
	if len(body) > maxConfigBytes {
		httpError(w, http.StatusRequestEntityTooLarge, "config is larger than 1 MiB")
		return
	}

	normalized, err := normalizeJSON(body)
	if err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	warnings, err := jsonparams.ValidateJSON(normalized)
	if err != nil {
		httpError(w, http.StatusBadRequest, "config validation failed: "+err.Error())
		return
	}

	err = os.MkdirAll(filepath.Dir(h.configPath), os.FileMode(0o777))
	if err != nil {
		httpError(w, http.StatusInternalServerError, "creating config directory: "+err.Error())
		return
	}

	if existing, readErr := os.ReadFile(h.configPath); readErr == nil {
		_ = os.WriteFile(h.configPath+".bak", existing, os.FileMode(0o666))
	}

	err = os.WriteFile(h.configPath, append(normalized, '\n'), os.FileMode(0o666))
	if err != nil {
		httpError(w, http.StatusInternalServerError, "writing config: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, configResponse{
		Path:        h.configPath,
		EnvConfig:   false,
		Config:      normalized,
		RestartHint: "Restart the container after saving so the updater reloads data/config.json.",
		Warnings:    warnings,
	})
}

func normalizeJSON(data []byte) (json.RawMessage, error) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	if _, ok := decoded["settings"]; !ok {
		decoded["settings"] = []any{}
	}
	settings, ok := decoded["settings"].([]any)
	if !ok {
		return nil, errors.New(`"settings" must be an array`)
	}
	decoded["settings"] = settings

	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(decoded); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buffer.Bytes()), nil
}

func appendValidationWarning(warnings []string, err error) []string {
	if err == nil {
		return warnings
	}
	return append(warnings, "Current config validation failed: "+err.Error())
}
