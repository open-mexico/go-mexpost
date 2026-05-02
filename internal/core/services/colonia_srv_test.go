package services_test

import (
	"errors"
	"testing"

	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/services"
	"github.com/stretchr/testify/assert"
)

// Mock del Repositorio
type MockRepo struct {
	SimularError bool
	DatosFalsos  []domain.Colonia
}

func (m *MockRepo) Buscar(filtroTipo string, valor string) ([]domain.Colonia, error) {
	if m.SimularError {
		return nil, errors.New("error en BD")
	}
	return m.DatosFalsos, nil
}

func TestBuscarColonias_ErroresDeValidacion(t *testing.T) {
	repo := &MockRepo{}
	servicio := services.NewColoniaService(repo)

	// Prueba 1: Sin parámetros
	_, err := servicio.BuscarColonias("", "")
	assert.Error(t, err)
	assert.Equal(t, "debes proporcionar un código postal o un nombre de colonia", err.Error())

	// Prueba 2: CP muy corto
	_, err = servicio.BuscarColonias("12", "")
	assert.Error(t, err)
	assert.Equal(t, "el código postal debe tener al menos 3 caracteres para la búsqueda parcial", err.Error())
}

func TestBuscarColonias_ExitoPorCP(t *testing.T) {
	mockDatos := []domain.Colonia{{Codigo: "06700", Nombre: "Roma Norte"}}
	repo := &MockRepo{DatosFalsos: mockDatos}
	servicio := services.NewColoniaService(repo)

	resultado, err := servicio.BuscarColonias("06700", "")
	assert.NoError(t, err)
	assert.Len(t, resultado, 1)
	assert.Equal(t, "Roma Norte", resultado[0].Nombre)
}