//go:build linux

package uname

import "syscall"

// Run calls syscall.Uname and returns the result.
func Run() (UnameResult, error) {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		return UnameResult{}, err
	}
	return UnameResult{
		Sysname:  charsToString(u.Sysname),
		Nodename: charsToString(u.Nodename),
		Release:  charsToString(u.Release),
		Version:  charsToString(u.Version),
		Machine:  charsToString(u.Machine),
	}, nil
}
