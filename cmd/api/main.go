package main

import (
	"log"
	"os"
	"strconv"

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

	// Rate limiting por IP: RATE_LIMIT req/s sostenidos, burst de RATE_BURST.
	// Valores por defecto: 10 req/s, burst 30.
	rateLimit := parseEnvFloat("RATE_LIMIT", 10.0)
	rateBurst := parseEnvInt("RATE_BURST", 30)
	router.Use(handler.RateLimitMiddleware(rateLimit, rateBurst))
	log.Printf("🛡️  Rate limit: %.0f req/s por IP, burst %d", rateLimit, rateBurst)

	router.GET("/colonias", apiHandler.BuscarColonias)
	router.GET("/colonias/id/:codigo_id", apiHandler.BuscarColoniaPorID)
	router.GET("/colonias/cercanas", apiHandler.BuscarColoniasCercanas)
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

func parseEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			return f
		}
	}
	return fallback
}

func parseEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
