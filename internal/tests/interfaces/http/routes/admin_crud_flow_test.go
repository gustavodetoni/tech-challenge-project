package routes_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	adminApp "github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	jwtAuth "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/routes"
	memrepo "github.com/soat-architecture/tech-challenge-project/internal/tests/infra/memory/repositories"
)

func TestAdminCRUD_Flow(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	store := memrepo.NewStore()
	userRepo := memrepo.NewUserRepository(store)
	clientRepo := memrepo.NewClientRepository(store)
	vehicleRepo := memrepo.NewVehicleRepository(store)
	serviceRepo := memrepo.NewServiceRepository(store)
	partRepo := memrepo.NewPartRepository(store)
	soRepo := memrepo.NewServiceOrderRepository(store)
	flowRepo := memrepo.NewServiceOrderFlowRepository(store, partRepo)

	jwtManager := jwtAuth.NewManager("secret", "issuer", "", 60*time.Minute)
	authMW := middlewares.NewAuthMiddleware(jwtManager)

	authService := appAuth.NewService(userRepo, jwtManager)

	clientSvc := adminApp.NewClientService(clientRepo)
	vehicleSvc := adminApp.NewVehicleService(vehicleRepo)
	serviceSvc := adminApp.NewServiceCatalogUseCase(serviceRepo)
	partSvc := adminApp.NewPartService(partRepo)
	serviceOrderSvc := adminApp.NewServiceOrderService(soRepo)
	userSvc := adminApp.NewUserService(userRepo)
	serviceOrderFlowSvc := serviceorder.NewFlowUseCase(clientRepo, vehicleRepo, serviceRepo, partRepo, flowRepo)

	router := gin.New()
	router.Use(gin.Recovery())
	routes.Register(router, routes.Deps{
		Health: controllers.NewHealthController(),
		Auth:   controllers.NewAuthController(authService, userRepo),
		AuthMW: authMW,

		AdminClients:       controllers.NewAdminClientsController(clientSvc),
		AdminVehicles:      controllers.NewAdminVehiclesController(vehicleSvc),
		AdminServices:      controllers.NewAdminServicesController(serviceSvc),
		AdminParts:         controllers.NewAdminPartsController(partSvc),
		AdminServiceOrders: controllers.NewAdminServiceOrdersController(serviceOrderSvc),
		AdminMetrics:       controllers.NewAdminMetricsController(serviceOrderSvc),
		AdminUsers:         controllers.NewAdminUsersController(userSvc),

		AdminServiceOrderFlow: controllers.NewAdminServiceOrderFlowController(serviceOrderFlowSvc),
		ClientServiceOrders:   controllers.NewClientServiceOrderController(serviceOrderFlowSvc),
	})

	adminToken, _, err := jwtManager.NewToken("admin-1", "ADMIN")
	require.NoError(t, err)

	// Clients CRUD
	var createdClient map[string]any
	doJSON(t, router, http.MethodPost, "/admin/clients", adminToken, map[string]any{
		"document_type":   "CPF",
		"document_number": "46420082412",
		"name":            "Maria",
	}, http.StatusCreated, &createdClient)
	clientID := createdClient["id"].(string)

	doJSON(t, router, http.MethodGet, "/admin/clients/"+clientID, adminToken, nil, http.StatusOK, nil)
	doJSON(t, router, http.MethodPut, "/admin/clients/"+clientID, adminToken, map[string]any{
		"document_type":   "CPF",
		"document_number": "46420082412",
		"name":            "Maria Updated",
	}, http.StatusOK, nil)
	doJSON(t, router, http.MethodGet, "/admin/clients?limit=10&offset=0", adminToken, nil, http.StatusOK, nil)

	// Vehicles CRUD (for client)
	var createdVehicle map[string]any
	doJSON(t, router, http.MethodPost, "/admin/clients/"+clientID+"/vehicles", adminToken, map[string]any{
		"plate":      "ABC1D23",
		"brand":      "Fiat",
		"model":      "Uno",
		"model_year": 2015,
	}, http.StatusCreated, &createdVehicle)
	vehicleID := createdVehicle["id"].(string)

	doJSON(t, router, http.MethodGet, "/admin/clients/"+clientID+"/vehicles?limit=10&offset=0", adminToken, nil, http.StatusOK, nil)
	doJSON(t, router, http.MethodGet, "/admin/vehicles/"+vehicleID, adminToken, nil, http.StatusOK, nil)
	doJSON(t, router, http.MethodPut, "/admin/vehicles/"+vehicleID, adminToken, map[string]any{
		"plate":      "ABC1D23",
		"brand":      "Fiat",
		"model":      "Uno Mille",
		"model_year": 2015,
	}, http.StatusOK, nil)

	// Services CRUD
	var createdService map[string]any
	doJSON(t, router, http.MethodPost, "/admin/services", adminToken, map[string]any{
		"name":              "Alinhamento",
		"base_price_cents":  15000,
		"estimated_minutes": 45,
		"active":            true,
	}, http.StatusCreated, &createdService)
	serviceID := createdService["id"].(string)
	doJSON(t, router, http.MethodGet, "/admin/services/"+serviceID, adminToken, nil, http.StatusOK, nil)
	doJSON(t, router, http.MethodPut, "/admin/services/"+serviceID, adminToken, map[string]any{
		"name":              "Alinhamento (updated)",
		"base_price_cents":  16000,
		"estimated_minutes": 50,
		"active":            true,
	}, http.StatusOK, nil)
	doJSON(t, router, http.MethodGet, "/admin/services?limit=10&offset=0", adminToken, nil, http.StatusOK, nil)

	// Parts CRUD + stock movement
	var createdPart map[string]any
	doJSON(t, router, http.MethodPost, "/admin/parts", adminToken, map[string]any{
		"sku":              "SKU-1",
		"name":             "Filtro",
		"unit_price_cents": 5000,
		"stock_quantity":   10,
		"active":           true,
	}, http.StatusCreated, &createdPart)
	partID := createdPart["id"].(string)
	doJSON(t, router, http.MethodGet, "/admin/parts/"+partID, adminToken, nil, http.StatusOK, nil)
	doJSON(t, router, http.MethodPost, "/admin/parts/"+partID+"/stock-movements", adminToken, map[string]any{
		"movement_type": "IN",
		"quantity":      5,
	}, http.StatusOK, nil)
	doJSON(t, router, http.MethodGet, "/admin/parts?limit=10&offset=0", adminToken, nil, http.StatusOK, nil)

	// Cleanup (delete)
	doJSON(t, router, http.MethodDelete, "/admin/vehicles/"+vehicleID, adminToken, nil, http.StatusNoContent, nil)
	doJSON(t, router, http.MethodDelete, "/admin/services/"+serviceID, adminToken, nil, http.StatusNoContent, nil)
	doJSON(t, router, http.MethodDelete, "/admin/parts/"+partID, adminToken, nil, http.StatusNoContent, nil)
	doJSON(t, router, http.MethodDelete, "/admin/clients/"+clientID, adminToken, nil, http.StatusNoContent, nil)
}
