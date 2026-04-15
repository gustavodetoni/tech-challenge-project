package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/domain/user"
)

func TestUser_RoleValues_AreStable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ADMIN", string(user.RoleAdmin))
	assert.Equal(t, "MANAGER", string(user.RoleManager))
	assert.Equal(t, "MECHANIC", string(user.RoleMechanic))
	assert.Equal(t, "ATTENDANT", string(user.RoleAttendant))
	assert.Equal(t, "VIEWER", string(user.RoleViewer))

	values := []user.Role{
		user.RoleAdmin,
		user.RoleManager,
		user.RoleMechanic,
		user.RoleAttendant,
		user.RoleViewer,
	}

	seen := map[string]struct{}{}
	for _, v := range values {
		_, exists := seen[string(v)]
		require.False(t, exists, "duplicate Role: %q", v)
		seen[string(v)] = struct{}{}
	}
}
