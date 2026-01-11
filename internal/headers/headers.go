package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	switch idx {
	case -1:
		return 0, false, nil
	case 0:
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	if len(parts) < 2 {
		return 0, false, fmt.Errorf("invalid header: missing ':'")
	}

	if len(parts[0]) == 0 {
		return 0, false, fmt.Errorf("invalid header: empty key")
	}

	if parts[0][len(parts[0])-1] == ' ' {
		return 0, false, fmt.Errorf("invalid header: space before ':'")
	}

	for _, c := range parts[0] {
		r := rune(c)
		if !isValidHeaderChar(r) {
			return 0, false, fmt.Errorf("invalid header key character (%v): must only contain alphabetical characters, digits, and special characters", r)
		}
	}

	key := bytes.ToLower(bytes.TrimSpace(parts[0]))
	value := bytes.TrimSpace(parts[1])

	if _, ok := h[string(key)]; ok {
		h[string(key)] += fmt.Sprintf(", %s", string(value))
	} else {
		h[string(key)] = string(value)
	}

	return idx + 2, false, nil
}

func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = value
}

func isValidHeaderChar(c rune) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' {
		return true
	}
	if '0' <= c && c <= '9' {
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}
