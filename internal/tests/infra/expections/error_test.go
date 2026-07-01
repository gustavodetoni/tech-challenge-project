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

func TestError_CodeOf_UnknownErrorReturnsInternal(t *testing.T) {
	t.Parallel()

	err := errors.New("boom")

	assert.False(t, expections.IsCode(err, expections.CodeInternal))
	assert.Equal(t, expections.CodeInternal, expections.CodeOf(err))
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
