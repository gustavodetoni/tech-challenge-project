package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
)

func TestClient_DocumentTypeValues_AreStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "CPF", string(client.DocumentTypeCPF))
	assert.Equal(t, "CNPJ", string(client.DocumentTypeCNPJ))

	values := []client.DocumentType{
		client.DocumentTypeCPF,
		client.DocumentTypeCNPJ,
	}

	seen := map[string]struct{}{}
	for _, v := range values {
		_, exists := seen[string(v)]
		require.False(t, exists, "duplicate DocumentType: %q", v)
		seen[string(v)] = struct{}{}
	}
}
