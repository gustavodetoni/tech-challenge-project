package shared_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
)

func TestParseLimitOffset_Defaults(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	limit, offset := shared.ParseLimitOffset(c)
	assert.Equal(t, 50, limit)
	assert.Equal(t, 0, offset)
}

func TestParseLimitOffset_ParsesQueryValues(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	r := httptest.NewRequest(http.MethodGet, "/x?limit=10&offset=5", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	limit, offset := shared.ParseLimitOffset(c)
	assert.Equal(t, 10, limit)
	assert.Equal(t, 5, offset)
}

func TestParseLimitOffset_IgnoresInvalidQueryValues(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	r := httptest.NewRequest(http.MethodGet, "/x?limit=nope&offset=bad", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = r

	limit, offset := shared.ParseLimitOffset(c)
	assert.Equal(t, 50, limit)
	assert.Equal(t, 0, offset)
}

func TestWriteRepoError_MapsNotFoundAndConflict(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		shared.WriteError(c, repository.ErrNotFound)
		assert.Equal(t, http.StatusNotFound, w.Code)
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		shared.WriteError(c, repository.ErrConflict)
		assert.Equal(t, http.StatusConflict, w.Code)
	}
}

func TestWriteRepoError_DefaultsToInternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	shared.WriteError(c, assert.AnError)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteRepoError_DelegatesToWriteError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	shared.WriteRepoError(c, repository.ErrNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWriteRepoError_MapsApplicationErrors(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "validation", err: expections.New(expections.CodeValidation, "invalid input"), want: http.StatusBadRequest},
		{name: "not found", err: expections.New(expections.CodeNotFound, "missing"), want: http.StatusNotFound},
		{name: "conflict", err: expections.New(expections.CodeConflict, "conflict"), want: http.StatusConflict},
		{name: "unauthorized", err: expections.New(expections.CodeUnauthorized, "unauthorized"), want: http.StatusUnauthorized},
		{name: "forbidden", err: expections.New(expections.CodeForbidden, "forbidden"), want: http.StatusForbidden},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			shared.WriteError(c, tt.err)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}
