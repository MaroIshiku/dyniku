package json

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MaroIshiku/dyniku/internal/models"
	"github.com/MaroIshiku/dyniku/pkg/publicip/ipversion"
)

const (
	publicIPLogDirPerm  = os.FileMode(0o777)
	publicIPLogFilePerm = os.FileMode(0o666)
	historyLineFields   = 2
)

// StoreNewIP stores a new IP address for a certain domain and owner.
func (db *Database) StoreNewIP(domain, owner string, ip netip.Addr, t time.Time) (err error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	targetIndex := -1
	for i, record := range db.data.Records {
		if record.Domain == domain && record.Owner == owner {
			targetIndex = i
			break
		}
	}

	recordNotFound := targetIndex == -1
	if recordNotFound {
		db.data.Records = append(db.data.Records, record{
			Domain: domain,
			Owner:  owner,
		})
		targetIndex = len(db.data.Records) - 1
	}

	event := models.HistoryEvent{
		IP:   ip,
		Time: t,
	}
	db.data.Records[targetIndex].Events = append(db.data.Records[targetIndex].Events, event)
	if err := db.write(); err != nil {
		return err
	}
	return db.writePublicIPLog(ip, t)
}

// GetEvents gets all the IP addresses history for a certain domain, owner and
// IP version, in the order from oldest to newest.
func (db *Database) GetEvents(domain, owner string,
	ipVersion ipversion.IPVersion,
) (events []models.HistoryEvent, err error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	for _, record := range db.data.Records {
		if record.Domain == domain && record.Owner == owner {
			return filterEvents(record.Events, ipVersion), nil
		}
	}
	return nil, nil
}

func (db *Database) writePublicIPLog(ip netip.Addr, t time.Time) error {
	if db.publicIPLogPath == "" || !ip.IsValid() {
		return nil
	}

	previousIP, err := lastLoggedIP(db.publicIPLogPath)
	if err != nil {
		return err
	}
	if previousIP == ip.String() {
		return nil
	}

	err = os.MkdirAll(filepath.Dir(db.publicIPLogPath), publicIPLogDirPerm)
	if err != nil {
		return fmt.Errorf("creating public IP log directory: %w", err)
	}

	line := t.UTC().Format("20060102-1504") + " " + ip.String() + "\n"
	file, err := os.OpenFile(db.publicIPLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, publicIPLogFilePerm)
	if err != nil {
		return fmt.Errorf("opening public IP log: %w", err)
	}
	_, writeErr := file.WriteString(line)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("writing public IP log: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing public IP log: %w", closeErr)
	}
	return nil
}

func lastLoggedIP(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("reading public IP log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return "", nil
	}

	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < historyLineFields {
		return "", nil
	}
	return fields[1], nil
}

func filterEvents(events []models.HistoryEvent, ipVersion ipversion.IPVersion) (filteredEvents []models.HistoryEvent) {
	filteredEvents = make([]models.HistoryEvent, 0, len(events))
	for _, event := range events {
		switch ipVersion {
		case ipversion.IP4:
			if event.IP.Is4() {
				filteredEvents = append(filteredEvents, event)
			}
		case ipversion.IP6:
			if event.IP.Is6() {
				filteredEvents = append(filteredEvents, event)
			}
		case ipversion.IP4or6:
			filteredEvents = append(filteredEvents, event)
		default:
			panic(fmt.Sprintf("IP version %v is not supported", ipVersion))
		}
	}
	return filteredEvents
}
