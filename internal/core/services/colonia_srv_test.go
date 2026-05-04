package services_test

import (
	"errors"
	"testing"

	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
	"github.com/open-mexico/go-mexpost/internal/core/services"
	"github.com/stretchr/testify/assert"
)

type MockRepo struct {
	Err error

	Colonias   []domain.Colonia
	Municipios []domain.Municipio

	LastColoniaFilter ports.ColoniaSearchFilter
	LastGeoFilter     ports.ReverseGeocodeFilter
}

func (m *MockRepo) SearchColonias(filter ports.ColoniaSearchFilter) ([]domain.Colonia, error) {
	m.LastColoniaFilter = filter
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Colonias, nil
}

func (m *MockRepo) SearchMunicipios(filter ports.MunicipioSearchFilter) ([]domain.Municipio, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	_ = filter
	return m.Municipios, nil
}

func (m *MockRepo) FindColoniasByPointBBox(filter ports.ReverseGeocodeFilter) ([]domain.Colonia, error) {
	m.LastGeoFilter = filter
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Colonias, nil
}

func TestBuscarColonias_ErroresDeValidacion(t *testing.T) {
	repo := &MockRepo{}
	servicio := services.NewColoniaService(repo)

	_, err := servicio.BuscarColonias(ports.ColoniaSearchFilter{}, false)
	assert.Error(t, err)
	assert.Equal(t, "debes proporcionar cp o nombre", err.Error())
}

func TestBuscarColonias_ExitoSinGeo(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[-99.2,19.3],[-99.1,19.3],[-99.1,19.4],[-99.2,19.4],[-99.2,19.3]]]}`
	mockDatos := []domain.Colonia{{Codigo: "06700", Nombre: "Roma Norte", Geometria: &geo}}
	repo := &MockRepo{Colonias: mockDatos}
	servicio := services.NewColoniaService(repo)

	resultado, err := servicio.BuscarColonias(ports.ColoniaSearchFilter{CP: "067"}, false)
	assert.NoError(t, err)
	assert.Len(t, resultado, 1)
	assert.Equal(t, "Roma Norte", resultado[0].Nombre)
	assert.Nil(t, resultado[0].Geometria)
	assert.Equal(t, "067", repo.LastColoniaFilter.CP)
}

func TestBuscarMunicipios_NotFound(t *testing.T) {
	repo := &MockRepo{}
	servicio := services.NewColoniaService(repo)

	_, err := servicio.BuscarMunicipios(ports.MunicipioSearchFilter{EstadoID: "09"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestBuscarPorCoordenadas_ValidaRangos(t *testing.T) {
	repo := &MockRepo{}
	servicio := services.NewColoniaService(repo)

	_, err := servicio.BuscarPorCoordenadas(ports.ReverseGeocodeFilter{Lat: 100, Lon: -99.1}, false)
	assert.Error(t, err)
	assert.Equal(t, "lat fuera de rango", err.Error())
}

func TestBuscarPorCoordenadas_PointInPolygon(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[-99.2,19.3],[-99.1,19.3],[-99.1,19.4],[-99.2,19.4],[-99.2,19.3]]]}`
	repo := &MockRepo{Colonias: []domain.Colonia{{Codigo: "06700", Nombre: "Roma Norte", Geometria: &geo}}}
	servicio := services.NewColoniaService(repo)

	resultado, err := servicio.BuscarPorCoordenadas(ports.ReverseGeocodeFilter{Lat: 19.35, Lon: -99.15}, true)
	assert.NoError(t, err)
	assert.NotNil(t, resultado)
	assert.Equal(t, "06700", resultado.Codigo)
}

func TestBuscarPorCoordenadas_NotFound(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[-99.2,19.3],[-99.1,19.3],[-99.1,19.4],[-99.2,19.4],[-99.2,19.3]]]}`
	repo := &MockRepo{Colonias: []domain.Colonia{{Codigo: "06700", Nombre: "Roma Norte", Geometria: &geo}}}
	servicio := services.NewColoniaService(repo)

	_, err := servicio.BuscarPorCoordenadas(ports.ReverseGeocodeFilter{Lat: 20.0, Lon: -99.15}, false)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestBuscarColonias_RepoError(t *testing.T) {
	repo := &MockRepo{Err: errors.New("error en bd")}
	servicio := services.NewColoniaService(repo)

	_, err := servicio.BuscarColonias(ports.ColoniaSearchFilter{CP: "067"}, false)
	assert.EqualError(t, err, "error en bd")
}
