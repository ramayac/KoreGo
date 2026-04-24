package dirname

import "testing"

func TestBasic(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/path/to/file.txt", "/path/to"},
		{"/path/to/dir", "/path/to"},
		{"file.txt", "."},
		{"/", "/"},
	}
	for _, c := range cases {
		got := Run(c.path).Result
		if got != c.want {
			t.Errorf("dirname(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}
