package configs

import (
	"strings"
	"testing"
)

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
		{name: "unknown field", input: strings.Replace(valid, `"auto_start"`, `"autoStart"`, 1), wantErr: true},
		{name: "trailing value", input: valid + `{}`, wantErr: true},
		{name: "null event", input: `{"events":[null]}`, wantErr: true},
		{name: "invalid time", input: strings.Replace(valid, `09:30`, `9:30`, 1), wantErr: true},
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
