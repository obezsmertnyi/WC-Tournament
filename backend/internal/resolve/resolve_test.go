package resolve

import "testing"

func TestNamesMatch(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Kylian Mbappé", "Kylian Mbappe", true},   // diacritics
		{"Mbappé", "Kylian Mbappé", true},          // partial (last name only)
		{"kylian mbappe", "Kylian MBAPPE", true},   // case
		{"Harry Kane", "Kylian Mbappé", false},     // different
		{"Lautaro Martínez", "Lionel Messi", false},
		{"", "Messi", false},
	}
	for _, c := range cases {
		got := namesMatch(normalizeName(c.a), normalizeName(c.b))
		if got != c.want {
			t.Errorf("namesMatch(%q,%q)=%v want %v", c.a, c.b, got, c.want)
		}
	}
}
