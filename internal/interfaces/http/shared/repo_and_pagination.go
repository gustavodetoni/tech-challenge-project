package shared

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
)

func ParseLimitOffset(c *gin.Context) (limit, offset int) {
	limit = 50
	offset = 0

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	return limit, offset
}

func WriteError(c *gin.Context, err error) {
	switch {
	case expections.IsCode(err, expections.CodeValidation):
		c.JSON(http.StatusBadRequest, middlewares.ErrorResponse(c, err.Error()))
	case expections.IsCode(err, expections.CodeNotFound):
		c.JSON(http.StatusNotFound, middlewares.ErrorResponse(c, "not found"))
	case expections.IsCode(err, expections.CodeConflict):
		c.JSON(http.StatusConflict, middlewares.ErrorResponse(c, "conflict"))
	case expections.IsCode(err, expections.CodeUnauthorized):
		c.JSON(http.StatusUnauthorized, middlewares.ErrorResponse(c, "unauthorized"))
	case expections.IsCode(err, expections.CodeForbidden):
		c.JSON(http.StatusForbidden, middlewares.ErrorResponse(c, "forbidden"))
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, middlewares.ErrorResponse(c, "not found"))
	case errors.Is(err, repository.ErrConflict):
		c.JSON(http.StatusConflict, middlewares.ErrorResponse(c, "conflict"))
	default:
		c.JSON(http.StatusInternalServerError, middlewares.ErrorResponse(c, "internal error"))
	}
}

func WriteHTTPError(c *gin.Context, status int, message string) {
	c.JSON(status, middlewares.ErrorResponse(c, message))
}

func WriteRepoError(c *gin.Context, err error) {
	WriteError(c, err)
}
