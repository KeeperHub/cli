package output

import "testing"

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short string unchanged", "hello", 10, "hello"},
		{"exact length unchanged", "hello", 5, "hello"},
		{"truncated with ellipsis", "hello world", 8, "hello..."},
		{"max too small returns original", "hello", 3, "hello"},
		{"max of 4 truncates", "hello", 4, "h..."},
		{"empty string", "", 10, ""},
		{"single char under max", "a", 10, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}
