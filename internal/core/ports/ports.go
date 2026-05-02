package ports

import "github.com/open-mexico/go-mexpost/internal/core/domain"

// ColoniaRepository define cómo la aplicación se comunica con la base de datos.
type ColoniaRepository interface {
	Buscar(filtroTipo string, valor string) ([]domain.Colonia, error)
}

// ColoniaService define las reglas de negocio accesibles desde la API.
type ColoniaService interface {
	BuscarColonias(codigoPostal string, nombre string) ([]domain.Colonia, error)
}