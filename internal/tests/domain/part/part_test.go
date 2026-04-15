package part

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
)

func TestPart_StockMovementTypeValues_AreStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "IN", string(part.StockMovementIn))
	assert.Equal(t, "OUT", string(part.StockMovementOut))
	assert.Equal(t, "ADJUSTMENT", string(part.StockMovementAdjustment))

	values := []part.StockMovementType{
		part.StockMovementIn,
		part.StockMovementOut,
		part.StockMovementAdjustment,
	}

	seen := map[string]struct{}{}
	for _, v := range values {
		_, exists := seen[string(v)]
		require.False(t, exists, "duplicate StockMovementType: %q", v)
		seen[string(v)] = struct{}{}
	}
}
