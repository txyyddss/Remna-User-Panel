package abuse

import "testing"

func TestNormalizeOutboundTags(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		valid bool
		want  []string
	}{
		{name: "trimmed", input: []string{" direct ", "proxy-main"}, valid: true, want: []string{"direct", "proxy-main"}},
		{name: "empty", input: nil},
		{name: "invalid character", input: []string{"direct", "proxy/tag"}},
		{name: "duplicate", input: []string{"direct", "DIRECT"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, valid := NormalizeOutboundTags(test.input)
			if valid != test.valid || len(got) != len(test.want) {
				t.Fatalf("NormalizeOutboundTags(%v) = %v, %t", test.input, got, valid)
			}
			for index := range got {
				if got[index] != test.want[index] {
					t.Fatalf("tag %d = %q, want %q", index, got[index], test.want[index])
				}
			}
		})
	}
}
