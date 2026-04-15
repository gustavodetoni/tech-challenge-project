package routes_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func doJSON(
	t *testing.T,
	h http.Handler,
	method string,
	path string,
	token string,
	body any,
	wantStatus int,
	out any,
) {
	t.Helper()

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	require.Equal(t, wantStatus, w.Code, "response body: %s", w.Body.String())

	if out != nil && w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), out))
	}
}
