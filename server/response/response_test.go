// +build unit

package response_test

import (
	"net/http"
	"testing"

	"github.com/fabiocicerchia/go-proxy-cache/server/response"
	"github.com/stretchr/testify/assert"
)

var MockStatusCode int
var MockContent [][]byte

type ResponseWriterMock struct {
	http.ResponseWriter
}

func (rwm ResponseWriterMock) WriteHeader(statusCode int) { MockStatusCode = statusCode }
func (rwm ResponseWriterMock) Write(p []byte) (int, error) {
	MockContent = append(MockContent, []byte{})
	chunk := len(MockContent) - 1
	MockContent[chunk] = append(MockContent[chunk], p...)
	return 0, nil
}

func TestNewWriter(t *testing.T) {
	var rwMock ResponseWriterMock

	lwr := response.NewLoggedResponseWriter(rwMock)

	assert.Equal(t, 0, lwr.StatusCode)
	assert.Len(t, lwr.Content, 0)

	tearDownResponse()
}

func TestCatchStatusCode(t *testing.T) {
	var rwMock ResponseWriterMock

	lwr := response.NewLoggedResponseWriter(rwMock)
	lwr.WriteHeader(http.StatusCreated)

	// checks lwr
	assert.Equal(t, http.StatusCreated, lwr.StatusCode)
	assert.Len(t, lwr.Content, 0)

	// verify calls on rwMock
	assert.Equal(t, http.StatusCreated, MockStatusCode)
	assert.Len(t, MockContent, 0)

	tearDownResponse()
}

func TestCatchContent(t *testing.T) {
	var rwMock ResponseWriterMock

	lwr := response.NewLoggedResponseWriter(rwMock)

	content := []byte("test content")
	_, err := lwr.Write(content)
	assert.Nil(t, err)

	expectedContent := [][]byte{content}

	// checks lwr
	assert.Equal(t, 0, lwr.StatusCode)
	assert.Equal(t, expectedContent, lwr.Content)

	// verify calls on rwMock
	assert.Equal(t, -1, MockStatusCode)
	assert.Equal(t, expectedContent, MockContent)

	tearDownResponse()
}

func tearDownResponse() {
	MockStatusCode = -1
	MockContent = make([][]byte, 0)
}
