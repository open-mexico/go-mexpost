package ports

import "github.com/open-mexico/go-mexpost/internal/core/domain"

// ColoniaSearchFilter define combinaciones de búsqueda para /colonias.
type ColoniaSearchFilter struct {
	CP               string
	Nombre           string
	MunicipioID      string
	MunicipioUID     string
	IncluirMunicipio bool
	SoloGeo          bool
	Limit            int
	Offset           int
}

const (
	DefaultLimit     = 100
	DefaultLimitGeo  = 50
	MaxLimit         = 500
	MaxLimitGeo      = 100
	DefaultNearLimit = 20
	MaxNearLimit     = 100
)

// ColoniaNearFilter define parámetros para /colonias/cercanas.
type ColoniaNearFilter struct {
	CP       string
	CodigoID string
	Limit    int
}

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
	FindColoniaByCodigoID(codigoID string) (*domain.Colonia, error)
	FindNearestColoniasByCodigoID(codigoID string, limit int) ([]domain.ColoniaCercana, error)
	FindNearestColoniasByCP(cp string, limit int) ([]domain.ColoniaCercana, error)
	SearchMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
	FindColoniasByPointBBox(filter ReverseGeocodeFilter) ([]domain.Colonia, error)
}

// ColoniaService define las reglas de negocio accesibles desde la API.
type ColoniaService interface {
	BuscarColonias(filter ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error)
	ContarColonias(filter ColoniaSearchFilter) (int, error)
	BuscarColoniaPorID(codigoID string, incluirGeo bool, incluirMunicipio bool) (*domain.Colonia, error)
	BuscarColoniasCercanas(filter ColoniaNearFilter, incluirGeo bool, incluirMunicipio bool) ([]domain.ColoniaCercana, error)
	BuscarMunicipios(filter MunicipioSearchFilter) ([]domain.Municipio, error)
	BuscarPorCoordenadas(filter ReverseGeocodeFilter, incluirGeo bool) (*domain.Colonia, error)
}
