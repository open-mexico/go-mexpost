package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/adapters/handler"
	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

// Mock del Servicio
type MockService struct {
	DevolverError bool
	Datos         []domain.Colonia
}

func (m *MockService) BuscarColonias(cp, nombre string) ([]domain.Colonia, error) {
	if m.DevolverError {
		return nil, errors.New("error de validación simulado")
	}
	return m.Datos, nil
}

func TestBuscarColonias_Status400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Preparamos el mock para que devuelva un error
	mockService := &MockService{DevolverError: true}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.Default()
	router.GET("/colonias", manejador.BuscarColonias)

	// Simulamos una petición web
	req, _ := http.NewRequest("GET", "/colonias?cp=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Validamos la respuesta HTTP
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error de validación simulado")
}