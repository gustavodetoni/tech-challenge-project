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
}

func Register(router *gin.Engine, deps Deps) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", deps.Health.Health)

	router.POST("/auth/register", deps.Auth.Register)
	router.POST("/auth/login", deps.Auth.Login)

	protected := router.Group("/")
	protected.Use(deps.AuthMW.RequireAuth())
	protected.GET("/me", deps.Auth.Me)
}
