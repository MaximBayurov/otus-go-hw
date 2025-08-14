package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	var b strings.Builder
	var prev string
	for _, v := range str {
		count, err := strconv.Atoi(string(v))
		isNumeric := err == nil
		if !isNumeric {
			if prev != "" {
				b.WriteString(prev)
			}
			prev = string(v)
			continue
		}
		if prev == "" {
			return "", ErrInvalidString
		}
		b.WriteString(strings.Repeat(prev, count))
		prev = ""
	}
	if prev != "" {
		b.WriteString(prev)
	}

	return b.String(), nil
}
