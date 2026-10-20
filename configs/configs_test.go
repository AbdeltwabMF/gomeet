package configs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseClockTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value     string
		wantHour  int
		wantMin   int
		wantValid bool
	}{
		{value: "09:30", wantHour: 9, wantMin: 30, wantValid: true},
		{value: "00:00", wantHour: 0, wantMin: 0, wantValid: true},
		{value: "23:59", wantHour: 23, wantMin: 59, wantValid: true},
		{value: "9:30", wantValid: false},
		{value: "+1:30", wantValid: false},
		{value: "-1:30", wantValid: false},
		{value: "1:300", wantValid: false},
		{value: "24:00", wantValid: false},
		{value: "12:60", wantValid: false},
		{value: "ab:cd", wantValid: false},
		{value: "", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			hour, minute, ok := ParseClockTime(tt.value)
			if ok != tt.wantValid {
				t.Fatalf("ParseClockTime(%q) valid = %v, want %v", tt.value, ok, tt.wantValid)
			}
			if ok && (hour != tt.wantHour || minute != tt.wantMin) {
				t.Fatalf("ParseClockTime(%q) = %d:%d, want %d:%d", tt.value, hour, minute, tt.wantHour, tt.wantMin)
			}
		})
	}
}

func TestDecodeConfig(t *testing.T) {
	t.Parallel()

	valid := `{
        "auto_start": true,
        "events": [{
            "summary": "Standup",
            "url": "https://meet.example.com/standup",
            "start": {"time": "09:30", "days": ["Monday", "Wednesday"]}
        }]
    }`
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: valid},
		{name: "null document", input: `null`, wantErr: true},
		{name: "unknown field", input: strings.Replace(valid, `"auto_start"`, `"autoStart"`, 1), wantErr: true},
		{name: "trailing value", input: valid + `{}`, wantErr: true},
		{name: "null event", input: `{"events":[null]}`, wantErr: true},
		{name: "invalid time", input: strings.Replace(valid, `09:30`, `9:30`, 1), wantErr: true},
		{name: "signed time", input: strings.Replace(valid, `09:30`, `+1:30`, 1), wantErr: true},
		{name: "invalid day", input: strings.Replace(valid, `Monday`, `Funday`, 1), wantErr: true},
		{name: "duplicate day", input: strings.Replace(valid, `"Monday", "Wednesday"`, `"Monday", "Monday"`, 1), wantErr: true},
		{name: "unsafe URL", input: strings.Replace(valid, `https://meet.example.com/standup`, `file:///tmp/meeting`, 1), wantErr: true},
		{name: "empty summary", input: strings.Replace(valid, `Standup`, ` `, 1), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := DecodeConfig(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecodeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsureConfigAtCreatesHelpfulEmptyTemplate(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.json")
	created, err := ensureConfigAt(path)
	if err != nil {
		t.Fatalf("ensureConfigAt() error = %v", err)
	}
	if !created {
		t.Fatal("ensureConfigAt() created = false, want true")
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open config: %v", err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close config: %v", err)
		}
	})
	cfg, err := DecodeConfig(f)
	if err != nil {
		t.Fatalf("starter config is invalid: %v", err)
	}
	if cfg.AutoStart || len(cfg.Events) != 0 {
		t.Fatalf("starter config = %+v, want disabled with no events", cfg)
	}

	created, err = ensureConfigAt(path)
	if err != nil || created {
		t.Fatalf("second ensureConfigAt() = (%v, %v), want (false, nil)", created, err)
	}
}
