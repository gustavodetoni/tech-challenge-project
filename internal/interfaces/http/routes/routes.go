package routes

import (
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/soat-architecture/tech-challenge-project/docs"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
)

type Deps struct {
	Health *controllers.HealthController
	Auth   *controllers.AuthController
	AuthMW *middlewares.AuthMiddleware

	AdminClients       *controllers.AdminClientsController
	AdminVehicles      *controllers.AdminVehiclesController
	AdminServices      *controllers.AdminServicesController
	AdminParts         *controllers.AdminPartsController
	AdminServiceOrders *controllers.AdminServiceOrdersController
	AdminMetrics       *controllers.AdminMetricsController
	AdminUsers         *controllers.AdminUsersController

	AdminServiceOrderFlow *controllers.AdminServiceOrderFlowController
	ClientServiceOrders   *controllers.ClientServiceOrderController
}

func Register(router *gin.Engine, deps Deps) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", deps.Health.Health)

	router.POST("/auth/register", deps.Auth.Register)
	router.POST("/auth/login", deps.Auth.Login)

	protected := router.Group("/")
	protected.Use(deps.AuthMW.RequireAuth())
	protected.GET("/me", deps.Auth.Me)

	adminGroup := router.Group("/admin")
	adminGroup.Use(deps.AuthMW.RequireRoles("ADMIN", "MANAGER"))

	adminGroup.GET("/clients", deps.AdminClients.List)
	adminGroup.POST("/clients", deps.AdminClients.Create)
	adminGroup.GET("/clients/:id", deps.AdminClients.Get)
	adminGroup.PUT("/clients/:id", deps.AdminClients.Update)
	adminGroup.DELETE("/clients/:id", deps.AdminClients.Delete)

	adminGroup.GET("/clients/:id/vehicles", deps.AdminVehicles.ListByClient)
	adminGroup.POST("/clients/:id/vehicles", deps.AdminVehicles.CreateForClient)
	adminGroup.GET("/vehicles/:id", deps.AdminVehicles.Get)
	adminGroup.PUT("/vehicles/:id", deps.AdminVehicles.Update)
	adminGroup.DELETE("/vehicles/:id", deps.AdminVehicles.Delete)

	adminGroup.GET("/services", deps.AdminServices.List)
	adminGroup.POST("/services", deps.AdminServices.Create)
	adminGroup.GET("/services/:id", deps.AdminServices.Get)
	adminGroup.PUT("/services/:id", deps.AdminServices.Update)
	adminGroup.DELETE("/services/:id", deps.AdminServices.Delete)

	adminGroup.GET("/parts", deps.AdminParts.List)
	adminGroup.POST("/parts", deps.AdminParts.Create)
	adminGroup.GET("/parts/:id", deps.AdminParts.Get)
	adminGroup.PUT("/parts/:id", deps.AdminParts.Update)
	adminGroup.DELETE("/parts/:id", deps.AdminParts.Delete)
	adminGroup.POST("/parts/:id/stock-movements", deps.AdminParts.AdjustStock)

	adminGroup.GET("/service-orders", deps.AdminServiceOrders.List)
	adminGroup.GET("/service-orders/:id", deps.AdminServiceOrders.Get)
	adminGroup.POST("/service-orders", deps.AdminServiceOrderFlow.CreateDraft)
	adminGroup.POST("/service-orders/:id/diagnosis/start", deps.AdminServiceOrderFlow.StartDiagnosis)
	adminGroup.POST("/service-orders/:id/budget/revise", deps.AdminServiceOrderFlow.ReviseBudget)
	adminGroup.POST("/service-orders/:id/budget/send", deps.AdminServiceOrderFlow.SendBudget)
	adminGroup.POST("/service-orders/:id/finish", deps.AdminServiceOrderFlow.Finish)
	adminGroup.POST("/service-orders/:id/deliver", deps.AdminServiceOrderFlow.Deliver)

	adminGroup.GET("/metrics/avg-execution-time", deps.AdminMetrics.AverageExecutionTime)

	adminOnly := router.Group("/admin")
	adminOnly.Use(deps.AuthMW.RequireRoles("ADMIN"))
	adminOnly.PUT("/users/:id/role", deps.AdminUsers.UpdateRole)

	clientGroup := router.Group("/client")
	clientGroup.GET("/service-orders/:code", deps.ClientServiceOrders.Get)
	clientGroup.POST("/service-orders/:code/budget/approve", deps.ClientServiceOrders.ApproveBudget)
	clientGroup.POST("/service-orders/:code/budget/reject", deps.ClientServiceOrders.RejectBudget)
}
