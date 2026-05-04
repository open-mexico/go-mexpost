package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/adapters/handler"
	"github.com/open-mexico/go-mexpost/internal/adapters/repository"
	"github.com/open-mexico/go-mexpost/internal/core/services"
)

func main() {
	// 1. Repositorio
	repo, err := repository.NewSQLiteRepository("./mapa.db")
	if err != nil {
		log.Fatalf("❌ Error: No se encontró la base de datos 'mapa.db'. Ejecuta 'go run ./cmd/setup/main.go' primero. Detalles: %v", err)
	}

	// 2. Servicios
	servicio := services.NewColoniaService(repo)

	// 3. Manejador HTTP
	apiHandler := handler.NewHttpHandler(servicio)

	// 4. Servidor Gin
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("❌ Error configurando proxies confiables: %v", err)
	}
	router.GET("/colonias", apiHandler.BuscarColonias)
	router.GET("/municipios", apiHandler.BuscarMunicipios)
	router.GET("/coordenadas", apiHandler.BuscarCoordenadas)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor corriendo en http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar servidor: %v", err)
	}
}
