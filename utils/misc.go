package utils

import (
	"time"
	"unicode/utf8"
)

func Sleep(milliseconds int64) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func FixInvalidUTF8(broken string) string {
	if !utf8.ValidString(broken) {
		v := make([]rune, 0, len(broken))
		for i, r := range broken {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(broken[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		broken = string(v)
	}
	return broken
}
