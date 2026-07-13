package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/adapters/handler"
	"github.com/open-mexico/go-mexpost/internal/adapters/repository"
	"github.com/open-mexico/go-mexpost/internal/core/services"
)

func main() {
	// Configure structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Repositorio
	repo, err := repository.NewSQLiteRepository("./mapa.db")
	if err != nil {
		logger.Error("No se encontró la base de datos 'mapa.db'. Ejecuta 'go run ./cmd/setup/main.go' primero", "error", err)
		os.Exit(1)
	}

	// 2. Servicios
	servicio := services.NewColoniaService(repo)

	// 3. Manejador HTTP
	apiHandler := handler.NewHttpHandler(servicio)

	// 4. Servidor Gin (New en lugar de Default para quitar el logger estándar)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(handler.SlogMiddleware(logger))

	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error("Error configurando proxies confiables", "error", err)
		os.Exit(1)
	}

	// CORS Config
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}
	corsConfig := cors.DefaultConfig()
	if corsOrigins == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = strings.Split(corsOrigins, ",")
	}
	router.Use(cors.New(corsConfig))

	// GZIP Config
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Rate limiting por IP: RATE_LIMIT req/s sostenidos, burst de RATE_BURST.
	rateLimit := parseEnvFloat("RATE_LIMIT", 10.0)
	rateBurst := parseEnvInt("RATE_BURST", 30)
	router.Use(handler.RateLimitMiddleware(rateLimit, rateBurst))
	logger.Info("Rate limit configurado", "req_per_sec", rateLimit, "burst", rateBurst)

	// Endpoints
	router.GET("/health", apiHandler.HealthCheck)
	router.GET("/colonias", apiHandler.BuscarColonias)
	router.GET("/colonias/id/:codigo_id", apiHandler.BuscarColoniaPorID)
	router.GET("/colonias/cercanas", apiHandler.BuscarColoniasCercanas)
	router.GET("/municipios", apiHandler.BuscarMunicipios)
	router.GET("/coordenadas", apiHandler.BuscarCoordenadas)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Iniciar servidor en una goroutine para permitir Graceful Shutdown
	go func() {
		logger.Info("Servidor corriendo", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error al iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	// Esperar señal de terminación (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Apagando servidor de manera gracefully...")

	// Dar hasta 10 segundos para terminar requests en vuelo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Servidor forzado a apagarse", "error", err)
	}

	logger.Info("Servidor detenido correctamente")
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
