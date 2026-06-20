package server

import (
	"context"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/qdm12/ddns-updater/internal/models"
	"github.com/qdm12/ddns-updater/internal/records"
	"github.com/qdm12/ddns-updater/pkg/publicip/ipversion"
)

type historyProvider struct{}

func (historyProvider) String() string { return "example" }
func (historyProvider) Domain() string { return "example.com" }
func (historyProvider) Owner() string  { return "@" }
func (historyProvider) BuildDomainName() string {
	return "example.com"
}
func (historyProvider) HTML() models.HTMLRow {
	return models.HTMLRow{}
}
func (historyProvider) Proxied() bool {
	return false
}
func (historyProvider) IPVersion() ipversion.IPVersion {
	return ipversion.IP4
}
func (historyProvider) IPv6Suffix() netip.Prefix {
	return netip.Prefix{}
}
func (historyProvider) Update(context.Context, *http.Client, netip.Addr) (netip.Addr, error) {
	return netip.Addr{}, nil
}

func TestMakeStatusRecordIncludesHistoryNewestFirst(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, time.June, 20, 8, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	record := records.Record{
		Provider: historyProvider{},
		History: models.History{
			{IP: netip.MustParseAddr("192.0.2.1"), Time: older},
			{IP: netip.MustParseAddr("192.0.2.2"), Time: newer},
		},
	}

	statusRecord := makeStatusRecord(record)

	if len(statusRecord.History) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(statusRecord.History))
	}
	if statusRecord.History[0].IP != "192.0.2.2" {
		t.Fatalf("expected newest IP first, got %s", statusRecord.History[0].IP)
	}
	if statusRecord.History[1].IP != "192.0.2.1" {
		t.Fatalf("expected older IP second, got %s", statusRecord.History[1].IP)
	}
}
