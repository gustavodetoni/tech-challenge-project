package plate

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{" abc-1234 ", "ABC1234"},
		{"AAA 1A23", "AAA1A23"},
		{"aaa-1a23", "AAA1A23"},
	}

	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Fatalf("Normalize(%q) expected %q, got %q", tt.in, tt.want, got)
		}
	}
}
