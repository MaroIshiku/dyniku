package server

import (
	"context"
	"time"

	"github.com/qdm12/ddns-updater/internal/records"
)

type Database interface {
	SelectAll() (records []records.Record)
}

type UpdateForcer interface {
	ForceUpdate(ctx context.Context) (errors []error)
}

type Logger interface {
	Info(s string)
	Warn(s string)
	Error(s string)
}

type StatusRecord struct {
	Domain    string    `json:"domain"`
	Owner     string    `json:"owner"`
	Provider  string    `json:"provider"`
	IPVersion string    `json:"ip_version"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	CurrentIP string    `json:"current_ip,omitempty"`
	Since     time.Time `json:"since,omitempty"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
}
