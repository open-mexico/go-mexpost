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
			where = append(where, "nombre_normalizado LIKE ?")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "nombre_normalizado = ?")
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
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	filas, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	var resultados []domain.Colonia
	for filas.Next() {
		var col domain.Colonia
		var codigo, nombre, tipo, ciudad, zona, estadoID, municipioID sql.NullString
		var minLon, minLat, maxLon, maxLat sql.NullFloat64
		err := filas.Scan(
			&codigo,
			&nombre,
			&tipo,
			&ciudad,
			&zona,
			&estadoID,
			&municipioID,
			&col.Geometria,
			&minLon,
			&minLat,
			&maxLon,
			&maxLat,
			&col.CentroLon,
			&col.CentroLat,
		)
		if err != nil {
			return nil, err
		}
		if codigo.Valid {
			col.Codigo = codigo.String
		}
		if nombre.Valid {
			col.Nombre = nombre.String
		}
		if tipo.Valid {
			col.Tipo = tipo.String
		}
		if ciudad.Valid {
			col.Ciudad = ciudad.String
		}
		if zona.Valid {
			col.Zona = zona.String
		}
		if estadoID.Valid {
			col.EstadoID = estadoID.String
		}
		if municipioID.Valid {
			col.MunicipioID = municipioID.String
		}
		if minLon.Valid {
			col.MinLon = minLon.Float64
		}
		if minLat.Valid {
			col.MinLat = minLat.Float64
		}
		if maxLon.Valid {
			col.MaxLon = maxLon.Float64
		}
		if maxLat.Valid {
			col.MaxLat = maxLat.Float64
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
			where = append(where, "nombre_normalizado LIKE ?")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "nombre_normalizado = ?")
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
	defer func() { _ = filas.Close() }()

	var resultados []domain.Municipio
	for filas.Next() {
		var m domain.Municipio
		var id, nombre, estadoID sql.NullString
		if err := filas.Scan(&id, &nombre, &estadoID); err != nil {
			return nil, err
		}
		if id.Valid {
			m.ID = id.String
		}
		if nombre.Valid {
			m.Nombre = nombre.String
		}
		if estadoID.Valid {
			m.EstadoID = estadoID.String
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
	defer func() { _ = filas.Close() }()

	var resultados []domain.Colonia
	for filas.Next() {
		var col domain.Colonia
		var codigo, nombre, tipo, ciudad, zona, estadoID, municipioID sql.NullString
		var minLon, minLat, maxLon, maxLat sql.NullFloat64
		if err := filas.Scan(
			&codigo,
			&nombre,
			&tipo,
			&ciudad,
			&zona,
			&estadoID,
			&municipioID,
			&col.Geometria,
			&minLon,
			&minLat,
			&maxLon,
			&maxLat,
			&col.CentroLon,
			&col.CentroLat,
		); err != nil {
			return nil, err
		}
		if codigo.Valid {
			col.Codigo = codigo.String
		}
		if nombre.Valid {
			col.Nombre = nombre.String
		}
		if tipo.Valid {
			col.Tipo = tipo.String
		}
		if ciudad.Valid {
			col.Ciudad = ciudad.String
		}
		if zona.Valid {
			col.Zona = zona.String
		}
		if estadoID.Valid {
			col.EstadoID = estadoID.String
		}
		if municipioID.Valid {
			col.MunicipioID = municipioID.String
		}
		if minLon.Valid {
			col.MinLon = minLon.Float64
		}
		if minLat.Valid {
			col.MinLat = minLat.Float64
		}
		if maxLon.Valid {
			col.MaxLon = maxLon.Float64
		}
		if maxLat.Valid {
			col.MaxLat = maxLat.Float64
		}
		resultados = append(resultados, col)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

var _ ports.ColoniaRepository = (*sqliteRepo)(nil)

func (r *sqliteRepo) CountColonias(filter ports.ColoniaSearchFilter) (int, error) {
	base := "SELECT COUNT(*) FROM colonias"
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
			where = append(where, "nombre_normalizado LIKE ? ")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "nombre_normalizado = ? ")
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

	var total int
	if err := r.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
