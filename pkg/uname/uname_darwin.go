//go:build darwin

package uname

import "golang.org/x/sys/unix"

// bytesToString converts a null-terminated byte array ([256]byte on Darwin) to a string.
func bytesToString(b [256]byte) string {
	n := 0
	for ; n < len(b) && b[n] != 0; n++ {
	}
	return string(b[:n])
}

// Run calls unix.Uname and returns the result.
func Run() (UnameResult, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return UnameResult{}, err
	}
	return UnameResult{
		Sysname:  bytesToString(u.Sysname),
		Nodename: bytesToString(u.Nodename),
		Release:  bytesToString(u.Release),
		Version:  bytesToString(u.Version),
		Machine:  bytesToString(u.Machine),
	}, nil
}
