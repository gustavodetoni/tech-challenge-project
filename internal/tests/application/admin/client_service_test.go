package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestClientService_Create_NormalizesAndValidates(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(c *client.Client) bool {
		return c != nil &&
			c.ID != "" &&
			c.Name == "Maria" &&
			c.DocumentType == client.DocumentTypeCPF &&
			c.DocumentNumber == "46420082412"
	})).Return(nil).Once()

	out, err := svc.Create(context.Background(), client.Client{
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: " 464.200.824-12 ",
		Name:           " Maria ",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "46420082412", out.DocumentNumber)
	assert.Equal(t, "Maria", out.Name)
	repo.AssertExpectations(t)
}

func TestClientService_Create_InvalidDocument(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	_, err := svc.Create(context.Background(), client.Client{
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "123",
		Name:           "Maria",
	})
	require.Error(t, err)
}

func TestClientService_Create_InvalidCNPJ(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	_, err := svc.Create(context.Background(), client.Client{
		DocumentType:   client.DocumentTypeCNPJ,
		DocumentNumber: "123",
		Name:           "Maria",
	})
	require.Error(t, err)
}

func TestClientService_Create_InvalidDocumentType(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	_, err := svc.Create(context.Background(), client.Client{
		DocumentType:   client.DocumentType("NOPE"),
		DocumentNumber: "46420082412",
		Name:           "Maria",
	})
	require.Error(t, err)
}

func TestClientService_Update_CallsRepoAndFetches(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	repo.On("Update", mock.Anything, mock.MatchedBy(func(c *client.Client) bool {
		return c != nil && c.ID == "c1" && c.Name == "Maria"
	})).Return(nil).Once()
	repo.On("FindByID", mock.Anything, "c1").Return(&client.Client{ID: "c1", Name: "Maria"}, nil).Once()

	out, err := svc.Update(context.Background(), "c1", client.Client{
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "c1", out.ID)
	repo.AssertExpectations(t)
}

func TestClientService_Update_RepoError(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.ClientRepository)
	svc := admin.NewClientService(repo)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*client.Client")).Return(assert.AnError).Once()

	_, err := svc.Update(context.Background(), "c1", client.Client{
		DocumentType:   client.DocumentTypeCPF,
		DocumentNumber: "46420082412",
		Name:           "Maria",
	})
	require.Error(t, err)
	repo.AssertExpectations(t)
}
