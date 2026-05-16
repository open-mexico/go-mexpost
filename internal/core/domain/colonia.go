package domain

import "errors"

// ErrNotFound se usa cuando la búsqueda no arroja resultados.
var ErrNotFound = errors.New("no se encontraron resultados")

// ValidationError representa errores de reglas de negocio/entrada.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// Colonia representa la estructura principal de los datos espaciales y postales.
type Colonia struct {
	Codigo          string
	Nombre          string
	Tipo            string
	Ciudad          string
	Zona            string
	EstadoID        string
	MunicipioID     string
	MunicipioNombre *string
	Geometria       *string
	MinLon          float64
	MinLat          float64
	MaxLon          float64
	MaxLat          float64
	CentroLon       *float64
	CentroLat       *float64
}

// Municipio representa los municipios del catálogo nacional.
type Municipio struct {
	ID       string
	Nombre   string
	EstadoID string
}
