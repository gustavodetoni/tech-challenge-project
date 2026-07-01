package serviceorder_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repository "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/client"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/order"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/part"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/service"
	"github.com/soat-architecture/tech-challenge-project/internal/domain/vehicle"
	repomocks "github.com/soat-architecture/tech-challenge-project/internal/tests/interfaces/repository/mocks"
)

func TestServiceOrder_CreateDraft_InvalidDocument(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "123",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_CreateDraft_InvalidPlate(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "INVALID",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_CreateDraft_BuildsLinesAndCallsFlow(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)

	svc := serviceorder.NewService(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1", DocumentNumber: "46420082412"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "ABC1D23"}, nil).Once()

	serviceRepo.On("FindByIDs", mock.Anything, []string{"s1", "s1"}).Return([]service.Service{
		{ID: "s1", Name: "Troca", BasePriceCents: 1000},
	}, nil).Once()
	partRepo.On("FindByIDs", mock.Anything, []string{"p1"}).Return([]part.Part{
		{ID: "p1", Name: "Filtro", UnitPriceCents: 5000},
	}, nil).Once()

	flowRepo.On("CreateDraft", mock.Anything, mock.MatchedBy(func(p repository.CreateServiceOrderDraftParams) bool {
		return p.ServiceOrder.ClientID == "c1" &&
			p.ServiceOrder.VehicleID == "v1" &&
			p.ServiceOrder.Status == order.StatusReceived &&
			p.Budget.Status == order.BudgetStatusDraft &&
			p.Budget.TotalAmountCents == 1000*2+5000*1 &&
			len(p.BudgetServices) == 1 &&
			p.BudgetServices[0].Quantity == 2 &&
			len(p.BudgetParts) == 1 &&
			p.BudgetParts[0].Quantity == 1
	})).Return(&order.ServiceOrder{
		ID:        "so1",
		Code:      "OS-20260414-ABCDEF",
		ClientID:  "c1",
		VehicleID: "v1",
		Status:    order.StatusReceived,
		OpenedAt:  time.Now().UTC(),
	}, &order.Budget{
		ID:               "b1",
		ServiceOrderID:   "so1",
		Version:          1,
		Status:           order.BudgetStatusDraft,
		TotalAmountCents: 7000,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil).Once()

	out, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "464.200.824-12",
		VehiclePlate:         "abc1d23",
		Services: []serviceorder.ItemInput{
			{ID: "s1", Quantity: 1},
			{ID: "s1", Quantity: 1},
		},
		Parts: []serviceorder.ItemInput{
			{ID: "p1", Quantity: 1},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "so1", out.ServiceOrderID)
	assert.Equal(t, "b1", out.BudgetID)
	assert.Equal(t, int64(7000), out.TotalCents)

	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
	serviceRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_UpdatesClientContact(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)

	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		flowRepo,
	)

	existingClient := &client.Client{ID: "c1", DocumentNumber: "46420082412"}
	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(existingClient, nil).Once()

	email := "email@example.com"
	phone := "+5511999999999"
	clientRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *client.Client) bool {
		return c.ID == "c1" &&
			c.Email != nil && *c.Email == email &&
			c.Phone != nil && *c.Phone == phone &&
			!c.UpdatedAt.IsZero()
	})).Return(nil).Once()

	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "ABC1D23"}, nil).Once()

	flowRepo.On("CreateDraft", mock.Anything, mock.MatchedBy(func(p repository.CreateServiceOrderDraftParams) bool {
		return p.ServiceOrder.ClientID == "c1" &&
			p.ServiceOrder.VehicleID == "v1" &&
			p.Budget.TotalAmountCents == 0 &&
			len(p.BudgetServices) == 0 &&
			len(p.BudgetParts) == 0
	})).Return(&order.ServiceOrder{ID: "so1", Code: "OS-ANY"}, &order.Budget{ID: "b1", TotalAmountCents: 0, Status: order.BudgetStatusDraft}, nil).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "464.200.824-12",
		ClientEmail:          &email,
		ClientPhone:          &phone,
		VehiclePlate:         "ABC1D23",
	})
	require.NoError(t, err)

	vehicleRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_UpdatesVehicleDetails(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)

	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		flowRepo,
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1", DocumentNumber: "46420082412"}, nil).Once()

	existingVehicle := &vehicle.Vehicle{ID: "v1", ClientID: "c1", Plate: "ABC1D23"}
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(existingVehicle, nil).Once()

	year := 2020
	color := "Preto"
	vehicleRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *vehicle.Vehicle) bool {
		return v.ID == "v1" &&
			v.ManufactureYear != nil && *v.ManufactureYear == year &&
			v.Color != nil && *v.Color == color &&
			!v.UpdatedAt.IsZero()
	})).Return(nil).Once()

	flowRepo.On("CreateDraft", mock.Anything, mock.MatchedBy(func(p repository.CreateServiceOrderDraftParams) bool {
		return p.ServiceOrder.ClientID == "c1" &&
			p.ServiceOrder.VehicleID == "v1"
	})).Return(&order.ServiceOrder{ID: "so1", Code: "OS-ANY"}, &order.Budget{ID: "b1", TotalAmountCents: 0, Status: order.BudgetStatusDraft}, nil).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:     client.DocumentTypeCPF,
		ClientDocumentNumber:   "46420082412",
		VehiclePlate:           "ABC1D23",
		VehicleManufactureYear: &year,
		VehicleColor:           &color,
	})
	require.NoError(t, err)

	clientRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_ReviseBudget_RequiresItems(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.ReviseBudget(context.Background(), "so1", serviceorder.ReviseBudgetInput{}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_ReviseBudget_CallsRepo(t *testing.T) {
	t.Parallel()

	serviceRepo := new(repomocks.ServiceRepository)
	partRepo := new(repomocks.PartRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		serviceRepo,
		partRepo,
		flowRepo,
	)

	serviceRepo.On("FindByIDs", mock.Anything, []string{"s1"}).Return([]service.Service{
		{ID: "s1", Name: "Troca", BasePriceCents: 1000},
	}, nil).Once()
	partRepo.On("FindByIDs", mock.Anything, []string{"p1"}).Return([]part.Part{
		{ID: "p1", Name: "Filtro", UnitPriceCents: 5000},
	}, nil).Once()

	flowRepo.On("CreateBudgetRevision", mock.Anything, mock.MatchedBy(func(p repository.CreateBudgetRevisionParams) bool {
		return p.ServiceOrderID == "so1" &&
			p.Budget.Status == order.BudgetStatusDraft &&
			p.Budget.TotalAmountCents == 6000 &&
			len(p.BudgetServices) == 1 &&
			len(p.BudgetParts) == 1
	})).Return(&order.Budget{
		ID:               "b2",
		ServiceOrderID:   "so1",
		Version:          2,
		Status:           order.BudgetStatusDraft,
		TotalAmountCents: 6000,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, nil).Once()

	out, err := svc.ReviseBudget(context.Background(), "so1", serviceorder.ReviseBudgetInput{
		Services: []serviceorder.ItemInput{{ID: "s1", Quantity: 1}},
		Parts:    []serviceorder.ItemInput{{ID: "p1", Quantity: 1}},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "b2", out.BudgetID)
	assert.Equal(t, 2, out.Version)
	assert.Equal(t, int64(6000), out.TotalCents)

	serviceRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_ClientNotFound_ReturnsInvalidInput(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(
		clientRepo,
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return((*client.Client)(nil), repository.ErrNotFound).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
	clientRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_VehicleNotFound_ReturnsInvalidInput(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return((*vehicle.Vehicle)(nil), repository.ErrNotFound).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_VehicleBelongsToAnotherClient_Conflict(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c2"}, nil).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.ErrorIs(t, err, repository.ErrConflict)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
}

func TestServiceOrder_ReviseBudget_MissingServiceOrderID(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.ReviseBudget(context.Background(), "", serviceorder.ReviseBudgetInput{
		Services: []serviceorder.ItemInput{{ID: "s1", Quantity: 1}},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_ClientActions_NormalizeDocument(t *testing.T) {
	t.Parallel()

	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(
		clientRepo,
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		flowRepo,
	)

	flowRepo.On("GetClientViewByCode", mock.Anything, "C-1", "46420082412").Return(&repository.ClientServiceOrderView{Code: "C-1"}, nil).Once()
	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{
		ID:             "cl-1",
		DocumentNumber: "46420082412",
		Name:           "",
	}, nil).Once()
	flowRepo.On("ApproveLatestBudgetByCode", mock.Anything, "C-1", "46420082412", (*string)(nil)).Return(nil).Once()
	flowRepo.On("RejectLatestBudgetByCode", mock.Anything, "C-1", "46420082412", "too expensive").Return(nil).Once()

	_, err := svc.ClientGetByCode(context.Background(), "C-1", "464.200.824-12")
	require.NoError(t, err)
	require.NoError(t, svc.ClientApproveBudget(context.Background(), "C-1", "464.200.824-12"))
	require.NoError(t, svc.ClientRejectBudget(context.Background(), "C-1", "464.200.824-12", "too expensive"))
	clientRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_FlowPassthrough_StartSendFinishDeliver(t *testing.T) {
	t.Parallel()

	flowRepo := new(repomocks.ServiceOrderFlowRepository)
	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		flowRepo,
	)

	changedBy := "u1"
	flowRepo.On("StartDiagnosis", mock.Anything, "so1", &changedBy).Return(nil).Once()
	flowRepo.On("SendLatestBudget", mock.Anything, "so1", &changedBy).Return(nil).Once()
	flowRepo.On("Finish", mock.Anything, "so1", &changedBy).Return(nil).Once()
	flowRepo.On("Deliver", mock.Anything, "so1", &changedBy).Return(nil).Once()

	require.NoError(t, svc.StartDiagnosis(context.Background(), "so1", &changedBy))
	require.NoError(t, svc.SendBudget(context.Background(), "so1", &changedBy))
	require.NoError(t, svc.Finish(context.Background(), "so1", &changedBy))
	require.NoError(t, svc.Deliver(context.Background(), "so1", &changedBy))

	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_ReviseBudget_InvalidItem_ReturnsInvalidInput(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.ReviseBudget(context.Background(), "so1", serviceorder.ReviseBudgetInput{
		Services: []serviceorder.ItemInput{{ID: "s1", Quantity: 0}},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_ReviseBudget_ServiceNotFound(t *testing.T) {
	t.Parallel()

	serviceRepo := new(repomocks.ServiceRepository)
	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		serviceRepo,
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	serviceRepo.On("FindByIDs", mock.Anything, []string{"s1"}).Return([]service.Service{}, nil).Once()

	_, err := svc.ReviseBudget(context.Background(), "so1", serviceorder.ReviseBudgetInput{
		Services: []serviceorder.ItemInput{{ID: "s1", Quantity: 1}},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	serviceRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_UsesExistingClientAndVehicle(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	flowRepo := new(repomocks.ServiceOrderFlowRepository)

	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		flowRepo,
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return(&vehicle.Vehicle{ID: "v1", ClientID: "c1"}, nil).Once()

	flowRepo.On("CreateDraft", mock.Anything, mock.MatchedBy(func(p repository.CreateServiceOrderDraftParams) bool {
		return p.ServiceOrder.ClientID == "c1" &&
			p.ServiceOrder.VehicleID == "v1" &&
			p.Budget.Status == order.BudgetStatusDraft &&
			len(p.BudgetServices) == 0 &&
			len(p.BudgetParts) == 0
	})).Return(
		&order.ServiceOrder{ID: "so1", Code: "C-1", ClientID: "c1", VehicleID: "v1", Status: order.StatusReceived, OpenedAt: time.Now().UTC()},
		&order.Budget{ID: "b1", ServiceOrderID: "so1", Version: 1, Status: order.BudgetStatusDraft, TotalAmountCents: 0, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		nil,
	).Once()

	out, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "464.200.824-12",
		VehiclePlate:         "ABC1D23",
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "so1", out.ServiceOrderID)
	assert.Equal(t, "b1", out.BudgetID)

	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
	flowRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_InvalidCNPJ(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCNPJ,
		ClientDocumentNumber: "123",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_CreateDraft_InvalidDocumentType(t *testing.T) {
	t.Parallel()

	svc := serviceorder.NewService(
		new(repomocks.ClientRepository),
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentType("NOPE"),
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, serviceorder.ErrInvalidInput)
}

func TestServiceOrder_CreateDraft_ClientFindError_Propagates(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	svc := serviceorder.NewService(
		clientRepo,
		new(repomocks.VehicleRepository),
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return((*client.Client)(nil), assert.AnError).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	clientRepo.AssertExpectations(t)
}

func TestServiceOrder_CreateDraft_VehicleFindError_Propagates(t *testing.T) {
	t.Parallel()

	clientRepo := new(repomocks.ClientRepository)
	vehicleRepo := new(repomocks.VehicleRepository)
	svc := serviceorder.NewService(
		clientRepo,
		vehicleRepo,
		new(repomocks.ServiceRepository),
		new(repomocks.PartRepository),
		new(repomocks.ServiceOrderFlowRepository),
	)

	clientRepo.On("FindByDocument", mock.Anything, "46420082412").Return(&client.Client{ID: "c1"}, nil).Once()
	vehicleRepo.On("FindByPlate", mock.Anything, "ABC1D23").Return((*vehicle.Vehicle)(nil), assert.AnError).Once()

	_, err := svc.CreateDraft(context.Background(), serviceorder.CreateDraftInput{
		ClientDocumentType:   client.DocumentTypeCPF,
		ClientDocumentNumber: "46420082412",
		VehiclePlate:         "ABC1D23",
	})
	require.Error(t, err)
	clientRepo.AssertExpectations(t)
	vehicleRepo.AssertExpectations(t)
}
