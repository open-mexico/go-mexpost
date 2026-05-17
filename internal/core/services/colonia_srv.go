package services

import (
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

func (s *coloniaService) BuscarColonias(filter ports.ColoniaSearchFilter, incluirGeo bool) ([]domain.Colonia, error) {
	filter.CP = strings.TrimSpace(filter.CP)
	filter.Nombre = strings.ToLower(strings.TrimSpace(filter.Nombre))
	filter.MunicipioID = strings.TrimSpace(filter.MunicipioID)
	filter.MunicipioUID = strings.TrimSpace(filter.MunicipioUID)

	if filter.CP == "" && filter.Nombre == "" {
		return nil, domain.ValidationError{Message: "debes proporcionar cp o nombre"}
	}

	if filter.CP != "" {
		if len(filter.CP) < 3 || len(filter.CP) > 5 {
			return nil, domain.ValidationError{Message: "cp invalido: usa entre 3 y 5 digitos (ej. 067 o 06700)"}
		}
		if !isOnlyDigits(filter.CP) {
			return nil, domain.ValidationError{Message: "cp invalido: solo se permiten digitos"}
		}
	}

	defaultLimit := ports.DefaultLimit
	maxLimit := ports.MaxLimit
	if incluirGeo {
		defaultLimit = ports.DefaultLimitGeo
		maxLimit = ports.MaxLimitGeo
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultLimit
	} else if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}

	resultados, err := s.repo.SearchColonias(filter)
	if err != nil {
		return nil, err
	}
	if len(resultados) == 0 {
		return nil, domain.ErrNotFound
	}

	if !incluirGeo {
		for i := range resultados {
			resultados[i].Geometria = nil
			resultados[i].CentroLat = nil
			resultados[i].CentroLon = nil
		}
	}

	if !filter.IncluirMunicipio {
		for i := range resultados {
			resultados[i].MunicipioNombre = nil
		}
	}

	return resultados, nil
}

func (s *coloniaService) BuscarColoniaPorID(codigoID string, incluirGeo bool, incluirMunicipio bool) (*domain.Colonia, error) {
	codigoID = strings.TrimSpace(codigoID)
	if codigoID == "" {
		return nil, domain.ValidationError{Message: "codigo_id es obligatorio"}
	}

	colonia, err := s.repo.FindColoniaByCodigoID(codigoID)
	if err != nil {
		return nil, err
	}
	if colonia == nil {
		return nil, domain.ErrNotFound
	}

	if !incluirGeo {
		colonia.Geometria = nil
		colonia.CentroLat = nil
		colonia.CentroLon = nil
	}

	if !incluirMunicipio {
		colonia.MunicipioNombre = nil
	}

	return colonia, nil
}

func (s *coloniaService) BuscarColoniasCercanas(filter ports.ColoniaNearFilter, incluirGeo bool, incluirMunicipio bool) ([]domain.ColoniaCercana, error) {
	filter.CP = strings.TrimSpace(filter.CP)
	filter.CodigoID = strings.TrimSpace(filter.CodigoID)

	if filter.CP == "" && filter.CodigoID == "" {
		return nil, domain.ValidationError{Message: "debes proporcionar cp o codigo_id"}
	}
	if filter.CP != "" && filter.CodigoID != "" {
		return nil, domain.ValidationError{Message: "usa solo uno: cp o codigo_id"}
	}

	if filter.CP != "" {
		if len(filter.CP) != 5 {
			return nil, domain.ValidationError{Message: "cp invalido: para cercania usa 5 digitos"}
		}
		if !isOnlyDigits(filter.CP) {
			return nil, domain.ValidationError{Message: "cp invalido: solo se permiten digitos"}
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = ports.DefaultNearLimit
	} else if filter.Limit > ports.MaxNearLimit {
		filter.Limit = ports.MaxNearLimit
	}

	var (
		resultados []domain.ColoniaCercana
		err        error
	)
	if filter.CodigoID != "" {
		resultados, err = s.repo.FindNearestColoniasByCodigoID(filter.CodigoID, filter.Limit)
	} else {
		resultados, err = s.repo.FindNearestColoniasByCP(filter.CP, filter.Limit)
	}
	if err != nil {
		return nil, err
	}
	if len(resultados) == 0 {
		return nil, domain.ErrNotFound
	}

	if !incluirGeo || !incluirMunicipio {
		for i := range resultados {
			if !incluirGeo {
				resultados[i].Colonia.Geometria = nil
				resultados[i].Colonia.CentroLat = nil
				resultados[i].Colonia.CentroLon = nil
			}
			if !incluirMunicipio {
				resultados[i].Colonia.MunicipioNombre = nil
			}
		}
	}

	return resultados, nil
}

func (s *coloniaService) BuscarMunicipios(filter ports.MunicipioSearchFilter) ([]domain.Municipio, error) {
	filter.Nombre = strings.ToLower(strings.TrimSpace(filter.Nombre))
	filter.EstadoID = strings.TrimSpace(filter.EstadoID)

	if filter.Nombre == "" && filter.EstadoID == "" {
		return nil, domain.ValidationError{Message: "debes proporcionar nombre o estado_id"}
	}

	resultados, err := s.repo.SearchMunicipios(filter)
	if err != nil {
		return nil, err
	}
	if len(resultados) == 0 {
		return nil, domain.ErrNotFound
	}

	return resultados, nil
}

func (s *coloniaService) BuscarPorCoordenadas(filter ports.ReverseGeocodeFilter, incluirGeo bool) (*domain.Colonia, error) {
	filter.EstadoID = strings.TrimSpace(filter.EstadoID)

	if filter.Lat < -90 || filter.Lat > 90 {
		return nil, domain.ValidationError{Message: "lat fuera de rango"}
	}
	if filter.Lon < -180 || filter.Lon > 180 {
		return nil, domain.ValidationError{Message: "lon fuera de rango"}
	}

	candidatas, err := s.repo.FindColoniasByPointBBox(filter)
	if err != nil {
		return nil, err
	}

	for i := range candidatas {
		geo := candidatas[i].Geometria
		if geo == nil || strings.TrimSpace(*geo) == "" {
			continue
		}

		inside, geoErr := PointInGeoJSON(filter.Lon, filter.Lat, *geo)
		if geoErr != nil {
			continue
		}
		if !inside {
			continue
		}

		if !incluirGeo {
			candidatas[i].Geometria = nil
			candidatas[i].CentroLat = nil
			candidatas[i].CentroLon = nil
		}

		return &candidatas[i], nil
	}

	return nil, domain.ErrNotFound
}

func (s *coloniaService) ContarColonias(filter ports.ColoniaSearchFilter) (int, error) {
	filter.CP = strings.TrimSpace(filter.CP)
	filter.Nombre = strings.ToLower(strings.TrimSpace(filter.Nombre))
	filter.MunicipioID = strings.TrimSpace(filter.MunicipioID)
	filter.MunicipioUID = strings.TrimSpace(filter.MunicipioUID)

	if filter.CP == "" && filter.Nombre == "" {
		return 0, domain.ValidationError{Message: "debes proporcionar cp o nombre"}
	}
	if filter.CP != "" {
		if len(filter.CP) < 3 || len(filter.CP) > 5 {
			return 0, domain.ValidationError{Message: "cp invalido: usa entre 3 y 5 digitos (ej. 067 o 06700)"}
		}
		if !isOnlyDigits(filter.CP) {
			return 0, domain.ValidationError{Message: "cp invalido: solo se permiten digitos"}
		}
	}

	return s.repo.CountColonias(filter)
}

func isOnlyDigits(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
