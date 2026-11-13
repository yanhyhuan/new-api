package helper

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiEmbeddingRequestValidation(t *testing.T) {
	for _, tc := range []struct {
		body         string
		batch, valid bool
	}{
		{`{"content":{"parts":[{"text":"ok"}]}}`, false, true},
		{`{"content":{"parts":[]}}`, false, false},
		{`{"requests":[]}`, true, false},
		{`{"requests":[null]}`, true, false},
		{`{"requests":[{"content":{"parts":[{"text":"ok"}]}}]}`, true, true},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
		c.Request.Header.Set("Content-Type", "application/json")
		var err error
		if tc.batch {
			_, err = GetAndValidateGeminiBatchEmbeddingRequest(c)
		} else {
			_, err = GetAndValidateGeminiEmbeddingRequest(c)
		}
		if tc.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
