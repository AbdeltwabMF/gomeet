package configs

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const LogFile = "log.txt"
const ConfigFile = "config.json"
const CredentialsFile = "credentials.json"
const TokenFile = "token.json"

type Config struct {
	AutoStart bool `json:"auto_start"`
}

type Start struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

type Event struct {
	Summary string `json:"summary"`
	Url     string `json:"url"`
	Start   Start  `json:"start"`
}

type Events struct {
	Items []*Event `json:"events"`
}

func LoadConfig() (*Config, error) {
	c, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(c, ConfigFile), os.O_CREATE|os.O_RDONLY, 0640)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func InitLogger() (*os.File, error) {
	lDir, err := platform.LogDir()
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(lDir, LogFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(file, nil))
	slog.SetDefault(logger)

	return file, nil
}
