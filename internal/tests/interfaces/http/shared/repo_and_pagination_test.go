package shared_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/shared"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/repository"
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
		shared.WriteRepoError(c, repository.ErrNotFound)
		assert.Equal(t, http.StatusNotFound, w.Code)
	}
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		shared.WriteRepoError(c, repository.ErrConflict)
		assert.Equal(t, http.StatusConflict, w.Code)
	}
}

func TestWriteRepoError_DefaultsToInternalError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	shared.WriteRepoError(c, assert.AnError)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
