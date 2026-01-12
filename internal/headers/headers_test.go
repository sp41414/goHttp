package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParser(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:9000\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:9000", headers.Get("Host"))
	assert.Equal(t, 22, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:9000  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character header
	headers = NewHeaders()
	data = []byte("H©st: localhost:9000\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple values for the same key
	headers = NewHeaders()
	data = []byte("Set-Language: python\r\nSet-Language: go\r\nSet-Language: typescript\r\n\r\n")
	n1, done, err := headers.Parse(data)
	require.NoError(t, err)
	n2, done, err := headers.Parse(data[n1:])
	require.NoError(t, err)
	n3, done, err := headers.Parse(data[n1+n2:])
	require.NoError(t, err)
	_, done, err = headers.Parse(data[n1+n2+n3:])
	require.NoError(t, err)
	require.True(t, done)
	require.NotNil(t, headers)
	assert.Equal(t, "python, go, typescript", headers.Get("Set-Language"))
	assert.Equal(t, 26, n3)
}
