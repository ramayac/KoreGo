package basename

import "testing"

func TestBasic(t *testing.T) {
	cases := []struct{ path, suffix, want string }{
		{"/path/to/file.txt", "", "file.txt"},
		{"/path/to/file.txt", ".txt", "file"},
		{"/path/to/dir/", "", "dir"},
		{"file", "", "file"},
		{"/", "", "/"},
	}
	for _, c := range cases {
		got := Run(c.path, c.suffix).Result
		if got != c.want {
			t.Errorf("basename(%q, %q) = %q, want %q", c.path, c.suffix, got, c.want)
		}
	}
}
