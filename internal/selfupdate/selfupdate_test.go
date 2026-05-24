package selfupdate

import "testing"

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
		wantErr            bool
	}{
		{"linux", "amd64", "ccmonitor-linux-amd64", false},
		{"linux", "arm64", "ccmonitor-linux-arm64", false},
		{"darwin", "arm64", "ccmonitor-darwin-arm64", false},
		{"darwin", "amd64", "ccmonitor-darwin-amd64", false},
		{"windows", "amd64", "", true},
		{"linux", "386", "", true},
	}
	for _, c := range cases {
		got, err := assetName(c.goos, c.goarch)
		if c.wantErr {
			if err == nil {
				t.Errorf("assetName(%q,%q) = %q, want error", c.goos, c.goarch, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("assetName(%q,%q) unexpected error: %v", c.goos, c.goarch, err)
		}
		if got != c.want {
			t.Errorf("assetName(%q,%q) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.3.0", "v1.4.0", -1},
		{"v1.4.0", "v1.3.0", 1},
		{"v1.4.0", "v1.4.0", 0},
		{"1.4.0", "v1.4.0", 0},   // leading v optional
		{"v1.4.0", "v1.4.1", -1}, // patch
		{"v1.10.0", "v1.9.0", 1}, // numeric, not lexical
		{"dev", "v1.4.0", -1},    // unparseable treated as 0.0.0
		{"v1.4.0-dirty", "v1.4.0", 0},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
