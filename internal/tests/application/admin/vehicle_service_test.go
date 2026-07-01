package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestVehicleService_Create_ValidatesPlateAndYears(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(v *vehicle.Vehicle) bool {
		return v != nil &&
			v.ID != "" &&
			v.ClientID == "c1" &&
			v.Plate == "ABC1D23"
	})).Return(nil).Once()

	_, err := svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "abc1d23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestVehicleService_CreateFromInput_MapsInput(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(v *vehicle.Vehicle) bool {
		return v != nil &&
			v.ClientID == "c1" &&
			v.Plate == "ABC1D23" &&
			v.Brand == "Fiat" &&
			v.Model == "Uno"
	})).Return(nil).Once()

	_, err := svc.CreateFromInput(context.Background(), admin.CreateVehicleInput{
		ClientID:  "c1",
		Plate:     "abc1d23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestVehicleService_Create_InvalidPlate(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	_, err := svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "INVALID",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.Error(t, err)
}

func TestVehicleService_Create_InvalidClientID(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	_, err := svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.Error(t, err)
}

func TestVehicleService_Create_RequiresBrandAndModel(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	_, err := svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "   ",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.Error(t, err)

	_, err = svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "   ",
		ModelYear: 2015,
	})
	require.Error(t, err)
}

func TestVehicleService_Create_InvalidModelYear(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	_, err := svc.Create(context.Background(), vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 1800,
	})
	require.Error(t, err)
}

func TestVehicleService_Update_RepoError(t *testing.T) {
	t.Parallel()

	repo := new(repomocks.VehicleRepository)
	svc := admin.NewVehicleService(repo)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*vehicle.Vehicle")).Return(assert.AnError).Once()

	_, err := svc.Update(context.Background(), "v1", vehicle.Vehicle{
		ClientID:  "c1",
		Plate:     "ABC1D23",
		Brand:     "Fiat",
		Model:     "Uno",
		ModelYear: 2015,
	})
	require.Error(t, err)
	repo.AssertExpectations(t)
}
