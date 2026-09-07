package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	adminApp "github.com/soat-architecture/tech-challenge-project/internal/application/admin"
	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	appPort "github.com/soat-architecture/tech-challenge-project/internal/application/port"
	"github.com/soat-architecture/tech-challenge-project/internal/application/serviceorder"
	"github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/email"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/controllers"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/middlewares"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/routes"
)

// @title Tech Challenge API
// @version 1.0
// @description API for the Tech Challenge project.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type `Bearer ` followed by your JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	gormDB, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")
	if err := db.InitAndCheckMigration(context.Background(), gormDB); err != nil {
		log.Fatal(err)
	}

	observabilityLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router := gin.New()
	router.Use(middlewares.RequestLogger(observabilityLogger))
	router.Use(middlewares.RequestRecovery(observabilityLogger))

	corsMW, err := config.NewCORSMiddleware(cfg.CORS)
	if err != nil {
		log.Fatal(err)
	}
	router.Use(corsMW)

	rateLimitMW, err := config.NewRateLimitMiddleware(cfg.RateLimit)
	if err != nil {
		log.Fatal(err)
	}
	router.Use(rateLimitMW)

	jwtManager := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, time.Duration(cfg.JWTExpiryMinutes)*time.Minute)
	clientJWTManager := auth.NewManager(cfg.JWTSecret, cfg.ClientJWTIssuer, cfg.ClientJWTAudience, time.Duration(cfg.JWTExpiryMinutes)*time.Minute)
	authMW := middlewares.NewAuthMiddlewareWithClientJWT(jwtManager, clientJWTManager)
	userRepo := repositories.NewUserRepository(gormDB)
	authService := appAuth.NewAuthUseCase(userRepo, jwtManager)

	clientRepo := repositories.NewClientRepository(gormDB)
	vehicleRepo := repositories.NewVehicleRepository(gormDB)
	serviceRepo := repositories.NewServiceRepository(gormDB)
	partRepo := repositories.NewPartRepository(gormDB)
	serviceOrderRepo := repositories.NewServiceOrderRepository(gormDB)
	serviceOrderFlowRepo := repositories.NewServiceOrderFlowRepository(gormDB)

	clientSvc := adminApp.NewClientAdminUseCase(clientRepo)
	vehicleSvc := adminApp.NewVehicleAdminUseCase(vehicleRepo)
	serviceSvc := adminApp.NewServiceCatalogUseCase(serviceRepo)
	partSvc := adminApp.NewPartInventoryUseCase(partRepo)
	serviceOrderSvc := adminApp.NewServiceOrderAdminUseCase(serviceOrderRepo)
	userSvc := adminApp.NewUserAdminUseCase(userRepo)
	var serviceOrderNotifier appPort.ServiceOrderNotifier
	if cfg.Brevo.APIKey != "" && cfg.Brevo.SenderEmail != "" {
		serviceOrderNotifier, err = email.NewBrevoNotifier(email.BrevoConfig{
			APIKey:      cfg.Brevo.APIKey,
			SenderEmail: cfg.Brevo.SenderEmail,
			SenderName:  cfg.Brevo.SenderName,
		}, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Brevo service order notifications enabled")
	}
	serviceOrderFlowSvc := serviceorder.NewServiceOrderFlowUseCase(clientRepo, vehicleRepo, serviceRepo, partRepo, serviceOrderFlowRepo, serviceOrderNotifier)

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

	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal(err)
	}
}
