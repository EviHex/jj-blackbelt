package doctor

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{{"0.43.0", "0.40.0", 1}, {"0.40.0", "0.40.0", 0}, {"0.39.9", "0.40.0", -1}, {"1.2", "1.2.0", 0}}
	for _, test := range tests {
		if got := compareVersions(test.left, test.right); got != test.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}
