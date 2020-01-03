package util

import (
	"github.com/bmatcuk/doublestar"
)

// Match returns whether the str matches the pattern
func Match(pattern, str string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}
	return doublestar.Match(pattern, str)
}
