package ports

import "github.com/open-mexico/go-mexpost/internal/core/domain"

// ColoniaSearchFilter define combinaciones de búsqueda para /colonias.
type ColoniaSearchFilter struct {
	CP          string
	Nombre      string
	MunicipioID string
	SoloGeo     bool
	Limit       int
	Offset      int
}

const (
	DefaultLimit    = 100
	DefaultLimitGeo = 50
	MaxLimit        = 500
	MaxLimitGeo     = 100
)

// MunicipioSearchFilter define combinaciones de búsqueda para /municipios.
type MunicipioSearchFilter struct {
	Nombre   string
	EstadoID string
}

// ReverseGeocodeFilter define parámetros para /coordenadas.
type ReverseGeocodeFilter struct {
	Lat      float64
	Lon      float64
	EstadoID string
}

// ColoniaRepository define cómo la aplicación se comunica con la base de datos.
type ColoniaRepository interface {
	SearchColonias(filter ColoniaSearchFilter) ([]domain.Colonia, error)
	CountColonias(filter ColoniaSearchFilter) (int, error)
	SearchMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
	FindColoniasByPointBBox(filter ReverseGeocodeFilter) ([]domain.Colonia, error)
}

// ColoniaService define las reglas de negocio accesibles desde la API.
type ColoniaService interface {
	BuscarColonias(filter ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error)
	ContarColonias(filter ColoniaSearchFilter) (int, error)
	BuscarMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
	BuscarPorCoordenadas(filter ReverseGeocodeFilter, incluirGeo bool) (*domain.Colonia, error)
}
