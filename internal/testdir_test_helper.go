package internal

import (
	"github.com/areknoster/hypert"
)

func WrapTestDataDir(t hypert.T, wraps int) string {
	wrap := func(prev func(t hypert.T) string) func(t hypert.T) string {
		return func(t hypert.T) string {
			return prev(t)
		}
	}
	wrapped := hypert.DefaultTestDataDir
	for i := 0; i < wraps; i++ {
		wrapped = wrap(wrapped)
	}
	return wrapped(t)
}
