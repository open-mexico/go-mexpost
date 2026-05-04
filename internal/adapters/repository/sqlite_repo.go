package repository

import (
	"database/sql"
	"strings"

	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepository(rutaDB string) (ports.ColoniaRepository, error) {
	db, err := sql.Open("sqlite", rutaDB)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &sqliteRepo{db: db}, nil
}

func (r *sqliteRepo) SearchColonias(filter ports.ColoniaSearchFilter) ([]domain.Colonia, error) {
	base := "SELECT codigo, nombre, tipo, ciudad, zona, estado_id, municipio_id, geometria, min_lon, min_lat, max_lon, max_lat, centro_lon, centro_lat FROM colonias"
	where := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.CP != "" {
		if len(filter.CP) >= 3 {
			where = append(where, "codigo LIKE ?")
			args = append(args, filter.CP+"%")
		} else {
			where = append(where, "codigo = ?")
			args = append(args, filter.CP)
		}
	}

	if filter.Nombre != "" {
		if len(filter.Nombre) >= 3 {
			where = append(where, "nombre LIKE ? COLLATE NOCASE")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "nombre = ? COLLATE NOCASE")
			args = append(args, filter.Nombre)
		}
	}

	if filter.MunicipioID != "" {
		where = append(where, "municipio_id = ?")
		args = append(args, filter.MunicipioID)
	}

	if filter.SoloGeo {
		where = append(where, "geometria IS NOT NULL AND TRIM(geometria) <> ''")
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY codigo, nombre"

	filas, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	var resultados []domain.Colonia
	for filas.Next() {
		var col domain.Colonia
		err := filas.Scan(
			&col.Codigo,
			&col.Nombre,
			&col.Tipo,
			&col.Ciudad,
			&col.Zona,
			&col.EstadoID,
			&col.MunicipioID,
			&col.Geometria,
			&col.MinLon,
			&col.MinLat,
			&col.MaxLon,
			&col.MaxLat,
			&col.CentroLon,
			&col.CentroLat,
		)
		if err != nil {
			return nil, err
		}
		resultados = append(resultados, col)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

func (r *sqliteRepo) SearchMunicipios(filter ports.MunicipioSearchFilter) ([]domain.Municipio, error) {
	base := "SELECT id, nombre, estado_id FROM municipios"
	where := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filter.Nombre != "" {
		if len(filter.Nombre) >= 3 {
			where = append(where, "nombre LIKE ? COLLATE NOCASE")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "nombre = ? COLLATE NOCASE")
			args = append(args, filter.Nombre)
		}
	}

	if filter.EstadoID != "" {
		where = append(where, "estado_id = ?")
		args = append(args, filter.EstadoID)
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY nombre"

	filas, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	var resultados []domain.Municipio
	for filas.Next() {
		var m domain.Municipio
		if err := filas.Scan(&m.ID, &m.Nombre, &m.EstadoID); err != nil {
			return nil, err
		}
		resultados = append(resultados, m)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

func (r *sqliteRepo) FindColoniasByPointBBox(filter ports.ReverseGeocodeFilter) ([]domain.Colonia, error) {
	query := `
		SELECT codigo, nombre, tipo, ciudad, zona, estado_id, municipio_id, geometria, min_lon, min_lat, max_lon, max_lat, centro_lon, centro_lat
		FROM colonias
		WHERE min_lat <= ?
		  AND max_lat >= ?
		  AND min_lon <= ?
		  AND max_lon >= ?`
	args := []any{filter.Lat, filter.Lat, filter.Lon, filter.Lon}

	if filter.EstadoID != "" {
		query += " AND estado_id = ?"
		args = append(args, filter.EstadoID)
	}

	query += " AND geometria IS NOT NULL AND TRIM(geometria) <> ''"
	query += " ORDER BY codigo, nombre"

	filas, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	var resultados []domain.Colonia
	for filas.Next() {
		var col domain.Colonia
		if err := filas.Scan(
			&col.Codigo,
			&col.Nombre,
			&col.Tipo,
			&col.Ciudad,
			&col.Zona,
			&col.EstadoID,
			&col.MunicipioID,
			&col.Geometria,
			&col.MinLon,
			&col.MinLat,
			&col.MaxLon,
			&col.MaxLat,
			&col.CentroLon,
			&col.CentroLat,
		); err != nil {
			return nil, err
		}
		resultados = append(resultados, col)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

var _ ports.ColoniaRepository = (*sqliteRepo)(nil)
