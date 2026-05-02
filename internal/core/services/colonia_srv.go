package services

import (
	"errors"
	"strings"

	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
)

type coloniaService struct {
	repo ports.ColoniaRepository
}

func NewColoniaService(repo ports.ColoniaRepository) ports.ColoniaService {
	return &coloniaService{repo: repo}
}

func (s *coloniaService) BuscarColonias(codigoPostal string, nombre string) ([]domain.Colonia, error) {
	cpLimpio := strings.TrimSpace(codigoPostal)
	nombreLimpio := strings.TrimSpace(nombre)

	// Reglas de negocio: Debe haber al menos un criterio de búsqueda válido
	if cpLimpio == "" && nombreLimpio == "" {
		return nil, errors.New("debes proporcionar un código postal o un nombre de colonia")
	}

	// Priorizamos la búsqueda por Código Postal
	if cpLimpio != "" {
		if len(cpLimpio) < 3 {
			return nil, errors.New("el código postal debe tener al menos 3 caracteres para la búsqueda parcial")
		}
		return s.repo.Buscar("codigo", cpLimpio)
	}

	// Si no hay CP, buscamos por Nombre
	if len(nombreLimpio) < 3 {
		return nil, errors.New("el nombre de la colonia debe tener al menos 3 caracteres")
	}
	return s.repo.Buscar("nombre", nombreLimpio)
}