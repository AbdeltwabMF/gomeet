package configs

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const (
	LogFile         = "log.txt"
	ConfigFile      = "config.json"
	CredentialsFile = "credentials.json"
	TokenFile       = "token.json"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
}

type start struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

type Event struct {
	Summary string `json:"summary"`
	Url     string `json:"url"`
	Start   start  `json:"start"`
}

type Events struct {
	Items []*Event `json:"events"`
}

type FuncInfo struct {
	Name string
	File string
	Line int
}

// OpenLog opens the log file with the specified flags.
func OpenLog(flags int) (*os.File, error) {
	d, err := platform.LogDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, LogFile), flags, 0640)
}

// OpenConfig opens the configuration file with the specified flags.
func OpenConfig(flags int) (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, ConfigFile), flags, 0640)
}

// OpenCredentials opens the credentials file with the specified flags.
func OpenCredentials(flags int) (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, CredentialsFile), flags, 0600)
}

// OpenToken opens the token file with the specified flags.
func OpenToken(flags int) (*os.File, error) {
	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(d, TokenFile), flags, 0600)
}

// LoadConfig loads the configuration from the configuration file.
func LoadConfig() (*Config, error) {
	f, err := OpenConfig(os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := new(Config)
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// InitLogger initializes the default logger with the provided writer.
func InitLogger(w io.Writer) {
	logger := slog.New(slog.NewTextHandler(w,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Any(a.Key, a.Value.Time().Format("2006-01-02 15:04:05"))
				}
				return a
			}},
	))
	slog.SetDefault(logger)
}

// CallerInfo returns information about the caller of the function where it's called.
func CallerInfo() FuncInfo {
	pc, file, line, ok := runtime.Caller(1) // 0: Function info, 1: Caller info
	if !ok {
		return FuncInfo{}
	}

	file = path.Base(file)
	name := path.Base(runtime.FuncForPC(pc).Name())

	return FuncInfo{Name: name, File: file, Line: line}
}
