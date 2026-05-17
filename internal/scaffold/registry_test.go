package scaffold

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"web", false},
		{"pwa", false},
		{"mobile", false},
		{"backend", false},
		{"ai", false},
		{"iot", false},
		{"nope", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			_, err := Resolve(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve(%q) err=%v, wantErr=%v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestAll_returnsAllPlatforms(t *testing.T) {
	all := All()
	if len(all) != 6 {
		t.Errorf("All() returned %d, want 6", len(all))
	}
}
