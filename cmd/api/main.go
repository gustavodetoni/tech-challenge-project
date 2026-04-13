package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	appAuth "github.com/soat-architecture/tech-challenge-project/internal/application/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db/repositories"
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

	router := gin.New()
	router.Use(gin.Recovery())

	jwtManager := auth.NewManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, time.Duration(cfg.JWTExpiryMinutes)*time.Minute)
	authMW := middlewares.NewAuthMiddleware(jwtManager)
	userRepo := repositories.NewUserRepository(gormDB)
	authService := appAuth.NewService(userRepo, jwtManager)

	routes.Register(router, routes.Deps{
		Health: controllers.NewHealthController(),
		Auth:   controllers.NewAuthController(authService, userRepo),
		AuthMW: authMW,
	})

	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal(err)
	}
}
