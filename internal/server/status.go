package server

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MaroIshiku/dyniku/internal/records"
)

type statusResponse struct {
	CurrentIP    string         `json:"current_ip,omitempty"`
	CurrentSince time.Time      `json:"current_since,omitzero"`
	Records      []StatusRecord `json:"records"`
	HistoryLog   []string       `json:"history_log"`
}

func (h *handlers) status(w http.ResponseWriter, _ *http.Request) {
	selectedIP := ""
	selectedIs4 := false
	var selectedSince time.Time
	records := h.db.SelectAll()
	statusRecords := make([]StatusRecord, 0, len(records))

	for _, record := range records {
		statusRecord := makeStatusRecord(record)
		statusRecords = append(statusRecords, statusRecord)

		currentIP := record.History.GetCurrentIP()
		if !currentIP.IsValid() {
			continue
		}
		if selectedIP == "" || (currentIP.Is4() && !selectedIs4) {
			selectedIP = currentIP.String()
			selectedSince = record.History.GetSuccessTime()
			selectedIs4 = currentIP.Is4()
		}
	}

	historyLog, err := readHistoryLog(h.publicIPLog)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "reading public IP log: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{
		CurrentIP:    selectedIP,
		CurrentSince: selectedSince,
		Records:      statusRecords,
		HistoryLog:   historyLog,
	})
}

func makeStatusRecord(record records.Record) StatusRecord {
	currentIP := record.History.GetCurrentIP()
	currentIPString := ""
	if currentIP.IsValid() {
		currentIPString = currentIP.String()
	}
	history := make([]IPEvent, 0, len(record.History))
	for i := len(record.History) - 1; i >= 0; i-- {
		event := record.History[i]
		if !event.IP.IsValid() {
			continue
		}
		history = append(history, IPEvent{
			IP:   event.IP.String(),
			Time: event.Time,
		})
	}
	return StatusRecord{
		Domain:    record.Provider.Domain(),
		Owner:     record.Provider.Owner(),
		Provider:  providerDisplayName(record.Provider.String()),
		IPVersion: record.Provider.IPVersion().String(),
		Status:    string(record.Status),
		Message:   record.Message,
		CurrentIP: currentIPString,
		Since:     record.History.GetSuccessTime(),
		CheckedAt: record.Time,
		History:   history,
	}
}

func providerDisplayName(providerString string) string {
	const marker = "provider:"
	lower := strings.ToLower(providerString)
	index := strings.Index(lower, marker)
	if index == -1 {
		return providerString
	}
	providerName := strings.TrimSpace(providerString[index+len(marker):])
	if separator := strings.Index(providerName, "|"); separator != -1 {
		providerName = strings.TrimSpace(providerName[:separator])
	}
	return providerName
}

func readHistoryLog(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}
