package common

import (
	"errors"
	"path/filepath"
	"strings"
)

// SecurePath resolves a target path against a base directory and ensures
// that the resulting path does not escape the base directory via ../ traversal or absolute paths.
// If baseDir is "/", all paths are allowed.
func SecurePath(target, baseDir string) (string, error) {
	if baseDir == "" {
		baseDir = "/"
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	absBase = filepath.Clean(absBase)

	var absTarget string
	if filepath.IsAbs(target) {
		absTarget = filepath.Clean(target)
	} else {
		absTarget = filepath.Clean(filepath.Join(absBase, target))
	}

	// Root base directory allows anything
	if absBase == "/" || absBase == filepath.VolumeName(absBase)+"\\" {
		return absTarget, nil
	}

	basePrefix := absBase
	if !strings.HasSuffix(basePrefix, string(filepath.Separator)) {
		basePrefix += string(filepath.Separator)
	}

	targetWithSep := absTarget
	if !strings.HasSuffix(targetWithSep, string(filepath.Separator)) {
		targetWithSep += string(filepath.Separator)
	}

	if absTarget != absBase && !strings.HasPrefix(targetWithSep, basePrefix) {
		return "", errors.New("path traversal detected: target escapes base directory")
	}

	return absTarget, nil
}
