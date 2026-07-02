package expections_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/infra/expections"
)

func TestError_New_ReturnsCodeAndMessage(t *testing.T) {
	t.Parallel()

	err := expections.New(expections.CodeValidation, "invalid name")

	require.Error(t, err)
	assert.Equal(t, "invalid name", err.Error())
	assert.True(t, expections.IsCode(err, expections.CodeValidation))
	assert.False(t, expections.IsCode(err, expections.CodeConflict))
	assert.Equal(t, expections.CodeValidation, expections.CodeOf(err))
}

func TestError_Wrap_PreservesCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("database conflict")
	err := expections.Wrap(expections.CodeConflict, "conflict", cause)

	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.True(t, expections.IsCode(err, expections.CodeConflict))
	assert.Equal(t, expections.CodeConflict, expections.CodeOf(err))
}

func TestError_ErrorFallsBackToCodeAndNilIsEmpty(t *testing.T) {
	t.Parallel()

	err := expections.New(expections.CodeNotFound, "")
	var nilErr *expections.Error

	assert.Equal(t, "NOT_FOUND", err.Error())
	assert.Equal(t, "", nilErr.Error())
}

func TestError_UnwrapNilError(t *testing.T) {
	t.Parallel()

	var err *expections.Error

	assert.NoError(t, err.Unwrap())
}

func TestError_CodeOf_UnknownErrorReturnsInternal(t *testing.T) {
	t.Parallel()

	err := errors.New("boom")

	assert.False(t, expections.IsCode(err, expections.CodeInternal))
	assert.Equal(t, expections.CodeInternal, expections.CodeOf(err))
}

func TestError_CodeOf_EmptyCodeReturnsInternal(t *testing.T) {
	t.Parallel()

	err := expections.New("", "missing code")

	assert.Equal(t, expections.CodeInternal, expections.CodeOf(err))
}

func TestError_ConstructorsUseExpectedCodes(t *testing.T) {
	t.Parallel()

	cause := errors.New("cause")

	assert.True(t, expections.IsCode(expections.Validation("invalid"), expections.CodeValidation))
	assert.True(t, expections.IsCode(expections.NotFound("missing", cause), expections.CodeNotFound))
	assert.True(t, expections.IsCode(expections.Conflict("conflict", cause), expections.CodeConflict))
}

func TestError_Internal_UsesCauseMessageWhenEmpty(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	err := expections.Internal("", cause)

	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
	assert.ErrorIs(t, err, cause)
	assert.True(t, expections.IsCode(err, expections.CodeInternal))
}
