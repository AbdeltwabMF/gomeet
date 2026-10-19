package platform

import "testing"

func TestValidateURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "HTTPS", url: "https://meet.example.com/room?key=value"},
		{name: "HTTP", url: "http://localhost:8080/room"},
		{name: "missing host", url: "https:///room", wantErr: true},
		{name: "unsafe scheme", url: "file:///tmp/meeting", wantErr: true},
		{name: "custom scheme", url: "zoommtg://example", wantErr: true},
		{name: "credentials", url: "https://user:pass@example.com", wantErr: true},
		{name: "surrounding whitespace", url: " https://example.com", wantErr: true},
		{name: "empty", url: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
