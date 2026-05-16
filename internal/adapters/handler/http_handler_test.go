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
	Err               error
	Colonias          []domain.Colonia
	Municipio         []domain.Municipio
	LastColoniaFilter ports.ColoniaSearchFilter
}

func (m *MockService) BuscarColonias(filter ports.ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error) {
	m.LastColoniaFilter = filter
	_ = incluirGeo
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Colonias, nil
}

func (m *MockService) ContarColonias(filter ports.ColoniaSearchFilter) (int, error) {
	_ = filter
	if m.Err != nil {
		return 0, m.Err
	}
	return len(m.Colonias), nil
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
	assert.Contains(t, w.Body.String(), "parametros invalidos")
	assert.Contains(t, w.Body.String(), "debes proporcionar cp o nombre")
}

func TestBuscarColonias_Status400CPInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockService{Err: domain.ValidationError{Message: "cp invalido: usa entre 3 y 5 digitos (ej. 067 o 06700)"}}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=14", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "parametros invalidos")
	assert.Contains(t, w.Body.String(), "cp invalido")
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
	assert.Contains(t, w.Body.String(), "\"total\"")
	assert.Contains(t, w.Body.String(), "\"pagina\"")
	assert.Contains(t, w.Body.String(), "\"total_paginas\"")
	assert.Contains(t, w.Body.String(), "\"pagina_anterior\"")
	assert.Contains(t, w.Body.String(), "\"pagina_siguiente\"")
}

func TestBuscarColonias_PaginacionMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	items := make([]domain.Colonia, 5)
	for i := range items {
		items[i] = domain.Colonia{Codigo: "06700", Nombre: "Roma"}
	}
	mockService := &MockService{Colonias: items}
	manejador := handler.NewHttpHandler(mockService)
	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	// Pagina 1 de 3 con limit=2, total=5 → pagina_siguiente existe, pagina_anterior nil
	req, _ := http.NewRequest("GET", "/colonias?cp=067&limit=2&pagina=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "\"pagina\":1")
	assert.Contains(t, body, "\"total_paginas\":3")
	assert.Contains(t, body, "pagina_siguiente")
	assert.Contains(t, body, "\"pagina_anterior\":null")
}

func TestBuscarColonias_Status400LimitInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manejador := handler.NewHttpHandler(&MockService{})
	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)
	req, _ := http.NewRequest("GET", "/colonias?cp=067&limit=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "limit debe ser un número entero positivo")
}

func TestBuscarColonias_Status400IncluirMunicipioInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manejador := handler.NewHttpHandler(&MockService{})
	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=067&incluir_municipio=talvez", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "incluir_municipio debe ser true o false")
}

func TestBuscarColonias_Status200ConMunicipio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	municipioNombre := "Cuauhtemoc"
	mockService := &MockService{Colonias: []domain.Colonia{{Codigo: "06700", Nombre: "Roma", MunicipioID: "015", EstadoID: "09", MunicipioNombre: &municipioNombre}}}
	manejador := handler.NewHttpHandler(mockService)

	router := gin.New()
	router.GET("/colonias", manejador.BuscarColonias)

	req, _ := http.NewRequest("GET", "/colonias?cp=067&incluir_municipio=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mockService.LastColoniaFilter.IncluirMunicipio)
	assert.Contains(t, w.Body.String(), "\"municipio_nombre\":\"Cuauhtemoc\"")
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
	assert.Contains(t, w.Body.String(), "ajusta tus filtros")
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
	assert.Contains(t, w.Body.String(), "ocurrio un error procesando la solicitud")
}
