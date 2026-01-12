// Package headers implements a parser and manager for HTTP-style header structures.
// It supports RFC-compliant character validation and case-insensitive key lookups.
package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// Headers is a key: value pair of strings which should be parsed from real headers
// or built from scratch.
type Headers map[string]string

// NewHeaders creates an empty Headers map.
func NewHeaders() Headers {
	return Headers{}
}

// Get retrieves the value for a key using a case-insensitive lookup.
func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

// Parse reads a single header line from the provided data.
// Lines should be formatted as:
//
//	Key: Value\r\n
//
// It returns the number of bytes read, a boolean 'done' which is true if
// an empty line (\r\n) is encountered, and any validation errors.
//
// If a duplicate key is found, the value is appended to the existing
// entry as a comma-separated list.
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

// OverrideValue sets the value for a key, replacing any existing data. The key is stored in lowercase.
func (h Headers) OverrideValue(key, value string) {
	h[strings.ToLower(key)] = value
}

// Override replaces prevKey with newKey. It returns true if prevKey existed and was successfully replaced.
func (h Headers) Override(prevKey, newKey, value string) bool {
	if _, ok := h[strings.ToLower(prevKey)]; ok {
		h[strings.ToLower(newKey)] = value
		delete(h, strings.ToLower(prevKey))
		return true
	} else {
		h[strings.ToLower(newKey)] = value
		return false
	}
}

// Add appends a value to a key. It validates the key against RFC 9110 tokens.
func (h Headers) Add(key, value string) error {
	for _, c := range key {
		r := rune(c)
		if !isValidHeaderChar(r) {
			return fmt.Errorf("invalid header key character (%v): must only contain alphabetical characters, digits, and special characters", r)
		}
	}

	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	if _, ok := h[string(key)]; ok {
		h[string(key)] += fmt.Sprintf(", %s", string(value))
	} else {
		h[string(key)] = string(value)
	}

	return nil
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
