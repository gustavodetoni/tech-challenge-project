package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/config"
	"github.com/soat-architecture/tech-challenge-project/internal/infra/db"
	"github.com/soat-architecture/tech-challenge-project/internal/interfaces/http/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.DatabaseURL != "" {
		if _, err := db.Connect(context.Background(), cfg.DatabaseURL); err != nil {
			log.Fatal(err)
		}
		log.Println("Connected to database")
	}

	router := gin.New()
	router.Use(gin.Recovery())

	routes.Register(router)

	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal(err)
	}
}
