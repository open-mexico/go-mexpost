package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/adapters/handler"
	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	Err       error
	Colonias  []domain.Colonia
	Municipio []domain.Municipio
}

func (m *MockService) BuscarColonias(filter ports.ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error) {
	_ = filter
	_ = incluirGeo
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Colonias, nil
}

func (m *MockService) BuscarMunicipios(filter ports.MunicipioSearchFilter) ([]domain.Municipio, error) {
	_ = filter
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Municipio, nil
}

func (m *MockService) BuscarPorCoordenadas(filter ports.ReverseGeocodeFilter, incluirGeo bool) (*domain.Colonia, error) {
	_ = filter
	_ = incluirGeo
	if m.Err != nil {
		return nil, m.Err
	}
	if len(m.Colonias) == 0 {
		return nil, domain.ErrNotFound
	}
	return &m.Colonias[0], nil
}

func TestBuscarColonias_Status400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockService{Err: domain.ValidationError{Message: "debes proporcionar cp o nombre"}}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "debes proporcionar cp o nombre")
}

func TestBuscarColonias_Status200SinGeo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	geo := `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`
	mockService := &MockService{Colonias: []domain.Colonia{{Codigo: "06700", Nombre: "Roma", Geometria: &geo}}}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=067", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"codigo\":\"06700\"")
	assert.NotContains(t, w.Body.String(), "geometria")
}

func TestBuscarMunicipios_Status404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockService{Err: domain.ErrNotFound}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/municipios", manejador.BuscarMunicipios)

	req, _ := http.NewRequest("GET", "/municipios?estado_id=09", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "no se encontraron resultados")
}

func TestBuscarCoordenadas_Status400PorParametros(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manejador := handler.NewHttpHandler(&MockService{})

	router := gin.New()
	router.GET("/coordenadas", manejador.BuscarCoordenadas)

	req, _ := http.NewRequest("GET", "/coordenadas?lat=19.4", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "lat y lon son obligatorios")
}

func TestBuscarCoordenadas_Status200ConGeo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	geo := `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`
	centroLon := -99.1
	centroLat := 19.4
	mockService := &MockService{Colonias: []domain.Colonia{{Codigo: "06700", Nombre: "Roma", Geometria: &geo, CentroLon: &centroLon, CentroLat: &centroLat}}}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/coordenadas", manejador.BuscarCoordenadas)

	req, _ := http.NewRequest("GET", "/coordenadas?lat=19.4&lon=-99.1&incluir_geo=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "geometria")
	assert.Contains(t, w.Body.String(), "centro_lon")
}

func TestBuscarColonias_Status500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &MockService{Err: errors.New("boom")}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=067", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error interno")
}
