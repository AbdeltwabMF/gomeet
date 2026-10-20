package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

const configTemplate = `{
  "auto_start": false,
  "events": []
}
`

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
	path, err := LogPath()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
}

// LogPath returns the path used for persistent logs.
func LogPath() (string, error) {
	d, err := platform.LogDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, LogFile), nil
}

// OpenConfig opens the configuration file for reading.
func OpenConfig() (*os.File, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}

// ConfigPath returns the path to the local calendar configuration.
func ConfigPath() (string, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, ConfigFile), nil
}

// EnsureConfig creates an empty starter configuration when none exists.
func EnsureConfig() (path string, created bool, err error) {
	path, err = ConfigPath()
	if err != nil {
		return "", false, err
	}

	created, err = ensureConfigAt(path)
	return path, created, err
}

func ensureConfigAt(path string) (bool, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if errors.Is(err, os.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("create starter config: %w", err)
	}

	if _, err := io.WriteString(f, configTemplate); err != nil {
		f.Close()
		os.Remove(path)
		return false, fmt.Errorf("write starter config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return false, fmt.Errorf("close starter config: %w", err)
	}

	return true, nil
}

// OpenCredentials opens the credentials file for reading.
func OpenCredentials() (*os.File, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}

// CredentialsPath returns the path to the optional Google credentials.
func CredentialsPath() (string, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, CredentialsFile), nil
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

	var cfg *Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("decode config: document must be a JSON object")
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
		if _, _, ok := ParseClockTime(e.Start.Time); !ok {
			return fmt.Errorf("event %d start time %q must use the 24-hour HH:MM format", i, e.Start.Time)
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

// ParseClockTime parses the strict 24-hour "15:04" format used for meeting
// start times. It rejects any other representation, including single digit
// hours and signed values.
func ParseClockTime(value string) (hour, minute int, ok bool) {
	if len(value) != len("15:04") || value[2] != ':' {
		return 0, 0, false
	}

	for _, i := range []int{0, 1, 3, 4} {
		if value[i] < '0' || value[i] > '9' {
			return 0, 0, false
		}
	}

	hour = int(value[0]-'0')*10 + int(value[1]-'0')
	minute = int(value[3]-'0')*10 + int(value[4]-'0')
	if hour > 23 || minute > 59 {
		return 0, 0, false
	}

	return hour, minute, true
}

// InitLogger initializes the default logger with the provided writer. Normal
// mode reports warnings and errors; verbose mode includes informational and
// diagnostic messages.
func InitLogger(w io.Writer, verbose bool) {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(w,
		&slog.HandlerOptions{
			Level:     level,
			AddSource: verbose,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Any(a.Key, a.Value.Time().Format("2006-01-02 15:04:05"))
				}
				return a
			}},
	))
	slog.SetDefault(logger)
}
