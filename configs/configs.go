package configs

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const (
	LogFile         = "log.txt"
	ConfigFile      = "config.json"
	CredentialsFile = "credentials.json"
	TokenFile       = "token.json"
)

type Config struct {
	AutoStart bool    `json:"auto_start"`
	Events    []Event `json:"events"`
}

type Start struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

type Event struct {
	Summary string `json:"summary"`
	URL     string `json:"url"`
	Start   Start  `json:"start"`
}

// OpenLog opens the log file for appending.
func OpenLog() (*os.File, error) {
	d, err := platform.LogDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, LogFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
}

// OpenConfig opens the configuration file for reading.
func OpenConfig() (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, ConfigFile), os.O_RDONLY, 0640)
}

// OpenCredentials opens the credentials file for reading.
func OpenCredentials() (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, CredentialsFile), os.O_RDONLY, 0600)
}

// OpenToken opens the token file for reading.
func OpenToken() (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, TokenFile), os.O_RDONLY, 0600)
}

// TokenDir returns the directory that stores the OAuth token.
func TokenDir() (string, error) {
	return platform.ConfigDir()
}

// HasCredentials reports whether Google Calendar credentials are available.
func HasCredentials() bool {
	d, err := platform.ConfigDir()
	if err != nil {
		return false
	}

	info, err := os.Stat(filepath.Join(d, CredentialsFile))

	return err == nil && !info.IsDir()
}

// LoadConfig loads the configuration from the configuration file.
func LoadConfig() (*Config, error) {
	f, err := OpenConfig()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return DecodeConfig(f)
}

// DecodeConfig decodes and validates one complete configuration document.
func DecodeConfig(r io.Reader) (*Config, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	cfg := new(Config)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode config: multiple JSON values")
		}
		return nil, fmt.Errorf("decode config trailing data: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks the configuration's semantic constraints.
func (cfg *Config) Validate() error {
	for i := range cfg.Events {
		e := &cfg.Events[i]
		if strings.TrimSpace(e.Summary) == "" {
			return fmt.Errorf("event %d summary must not be empty", i)
		}
		if err := platform.ValidateURL(e.URL); err != nil {
			return fmt.Errorf("event %d URL: %w", i, err)
		}
		if err := validateClockTime(e.Start.Time); err != nil {
			return fmt.Errorf("event %d start time %q: %w", i, e.Start.Time, err)
		}
		if len(e.Start.Days) == 0 {
			return fmt.Errorf("event %d must include at least one day", i)
		}

		seen := make(map[time.Weekday]bool, len(e.Start.Days))
		for _, day := range e.Start.Days {
			weekday, ok := parseWeekday(day)
			if !ok {
				return fmt.Errorf("event %d has invalid weekday %q", i, day)
			}
			if seen[weekday] {
				return fmt.Errorf("event %d repeats weekday %q", i, day)
			}
			seen[weekday] = true
		}
	}

	return nil
}

func parseWeekday(day string) (time.Weekday, bool) {
	for weekday := time.Sunday; weekday <= time.Saturday; weekday++ {
		if day == weekday.String() {
			return weekday, true
		}
	}

	return 0, false
}

// validateClockTime requires the exact 24-hour "15:04" form used for matching.
func validateClockTime(value string) error {
	if len(value) != len("15:04") || value[2] != ':' {
		return fmt.Errorf("must use the 24-hour HH:MM format")
	}

	hour, err := strconv.Atoi(value[:2])
	if err != nil || hour < 0 || hour > 23 {
		return fmt.Errorf("must be a valid 24-hour time")
	}
	minute, err := strconv.Atoi(value[3:])
	if err != nil || minute < 0 || minute > 59 {
		return fmt.Errorf("must be a valid 24-hour time")
	}

	return nil
}

// InitLogger initializes the default logger with the provided writer.
func InitLogger(w io.Writer) {
	logger := slog.New(slog.NewTextHandler(w,
		&slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Any(a.Key, a.Value.Time().Format("2006-01-02 15:04:05"))
				}
				return a
			}},
	))
	slog.SetDefault(logger)
}
