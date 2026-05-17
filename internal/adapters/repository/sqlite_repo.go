package repository

import (
	"database/sql"
	"math"
	"strings"

	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
	_ "modernc.org/sqlite"
)

type sqliteRepo struct {
	db *sql.DB
}

const coloniaViewSelect = `
	v.codigo_id,
	v.codigo,
	v.colonia_nombre,
	v.tipo,
	v.ciudad,
	v.zona,
	v.estado_id,
	v.municipio_id,
	v.municipio_uid,
	v.municipio_nombre,
	v.geometria,
	v.min_lon,
	v.min_lat,
	v.max_lon,
	v.max_lat,
	v.centro_lon,
	v.centro_lat`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanColoniaCercana(scanner rowScanner) (domain.ColoniaCercana, error) {
	var result domain.ColoniaCercana
	var codigoID, codigo, nombre, tipo, ciudad, zona, estadoID, municipioID sql.NullString
	var municipioUID, municipioNombre, geometria sql.NullString
	var minLon, minLat, maxLon, maxLat sql.NullFloat64
	var centroLon, centroLat sql.NullFloat64
	var dist2 sql.NullFloat64

	err := scanner.Scan(
		&codigoID,
		&codigo,
		&nombre,
		&tipo,
		&ciudad,
		&zona,
		&estadoID,
		&municipioID,
		&municipioUID,
		&municipioNombre,
		&geometria,
		&minLon,
		&minLat,
		&maxLon,
		&maxLat,
		&centroLon,
		&centroLat,
		&dist2,
	)
	if err != nil {
		return result, err
	}

	if codigoID.Valid {
		result.Colonia.CodigoID = codigoID.String
	}
	if codigo.Valid {
		result.Colonia.Codigo = codigo.String
	}
	if nombre.Valid {
		result.Colonia.Nombre = nombre.String
	}
	if tipo.Valid {
		result.Colonia.Tipo = tipo.String
	}
	if ciudad.Valid {
		result.Colonia.Ciudad = ciudad.String
	}
	if zona.Valid {
		result.Colonia.Zona = zona.String
	}
	if estadoID.Valid {
		result.Colonia.EstadoID = estadoID.String
	}
	if municipioID.Valid {
		result.Colonia.MunicipioID = municipioID.String
	}
	if municipioUID.Valid {
		result.Colonia.MunicipioUID = municipioUID.String
	}
	if municipioNombre.Valid {
		value := municipioNombre.String
		result.Colonia.MunicipioNombre = &value
	}
	if geometria.Valid {
		value := geometria.String
		result.Colonia.Geometria = &value
	}
	if minLon.Valid {
		result.Colonia.MinLon = minLon.Float64
	}
	if minLat.Valid {
		result.Colonia.MinLat = minLat.Float64
	}
	if maxLon.Valid {
		result.Colonia.MaxLon = maxLon.Float64
	}
	if maxLat.Valid {
		result.Colonia.MaxLat = maxLat.Float64
	}
	if centroLon.Valid {
		value := centroLon.Float64
		result.Colonia.CentroLon = &value
	}
	if centroLat.Valid {
		value := centroLat.Float64
		result.Colonia.CentroLat = &value
	}
	if dist2.Valid && dist2.Float64 > 0 {
		// Conversión aproximada de grados a km para exponer una métrica útil al cliente.
		result.DistanciaKM = math.Sqrt(dist2.Float64) * 111.32
	}

	return result, nil
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

func scanColonia(scanner rowScanner) (domain.Colonia, error) {
	var col domain.Colonia
	var codigoID, codigo, nombre, tipo, ciudad, zona, estadoID, municipioID sql.NullString
	var municipioUID, municipioNombre, geometria sql.NullString
	var minLon, minLat, maxLon, maxLat sql.NullFloat64
	var centroLon, centroLat sql.NullFloat64

	err := scanner.Scan(
		&codigoID,
		&codigo,
		&nombre,
		&tipo,
		&ciudad,
		&zona,
		&estadoID,
		&municipioID,
		&municipioUID,
		&municipioNombre,
		&geometria,
		&minLon,
		&minLat,
		&maxLon,
		&maxLat,
		&centroLon,
		&centroLat,
	)
	if err != nil {
		return col, err
	}

	if codigoID.Valid {
		col.CodigoID = codigoID.String
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
	if municipioUID.Valid {
		col.MunicipioUID = municipioUID.String
	}
	if municipioNombre.Valid {
		value := municipioNombre.String
		col.MunicipioNombre = &value
	}
	if geometria.Valid {
		value := geometria.String
		col.Geometria = &value
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
	if centroLon.Valid {
		value := centroLon.Float64
		col.CentroLon = &value
	}
	if centroLat.Valid {
		value := centroLat.Float64
		col.CentroLat = &value
	}

	return col, nil
}

func (r *sqliteRepo) SearchColonias(filter ports.ColoniaSearchFilter) ([]domain.Colonia, error) {
	base := "SELECT " + coloniaViewSelect + " FROM vw_colonias_busqueda v"
	where := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filter.CP != "" {
		if len(filter.CP) >= 3 {
			where = append(where, "v.codigo LIKE ?")
			args = append(args, filter.CP+"%")
		} else {
			where = append(where, "v.codigo = ?")
			args = append(args, filter.CP)
		}
	}

	if filter.Nombre != "" {
		if len(filter.Nombre) >= 3 {
			where = append(where, "v.colonia_nombre_normalizado LIKE ?")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "v.colonia_nombre_normalizado = ?")
			args = append(args, filter.Nombre)
		}
	}

	if filter.MunicipioID != "" {
		where = append(where, "v.municipio_id = ?")
		args = append(args, filter.MunicipioID)
	}

	if filter.MunicipioUID != "" {
		where = append(where, "v.municipio_uid = ?")
		args = append(args, filter.MunicipioUID)
	}

	if filter.SoloGeo {
		where = append(where, "v.geometria IS NOT NULL AND TRIM(v.geometria) <> ''")
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY v.codigo, v.colonia_nombre"
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
		col, err := scanColonia(filas)
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
	base := "SELECT id, nombre, estado_id, municipio_uid FROM municipios"
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
		var id, nombre, estadoID, municipioUID sql.NullString
		if err := filas.Scan(&id, &nombre, &estadoID, &municipioUID); err != nil {
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
		if municipioUID.Valid {
			m.MunicipioUID = municipioUID.String
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
		SELECT ` + coloniaViewSelect + `
		FROM vw_colonias_busqueda v
		WHERE v.min_lat <= ?
		  AND v.max_lat >= ?
		  AND v.min_lon <= ?
		  AND v.max_lon >= ?`
	args := []any{filter.Lat, filter.Lat, filter.Lon, filter.Lon}

	if filter.EstadoID != "" {
		query += " AND v.estado_id = ?"
		args = append(args, filter.EstadoID)
	}

	query += " AND v.geometria IS NOT NULL AND TRIM(v.geometria) <> ''"
	query += " ORDER BY v.codigo, v.colonia_nombre"

	filas, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	var resultados []domain.Colonia
	for filas.Next() {
		col, err := scanColonia(filas)
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

var _ ports.ColoniaRepository = (*sqliteRepo)(nil)

func (r *sqliteRepo) CountColonias(filter ports.ColoniaSearchFilter) (int, error) {
	base := "SELECT COUNT(*) FROM vw_colonias_busqueda"
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
			where = append(where, "colonia_nombre_normalizado LIKE ?")
			args = append(args, "%"+filter.Nombre+"%")
		} else {
			where = append(where, "colonia_nombre_normalizado = ?")
			args = append(args, filter.Nombre)
		}
	}
	if filter.MunicipioID != "" {
		where = append(where, "municipio_id = ?")
		args = append(args, filter.MunicipioID)
	}
	if filter.MunicipioUID != "" {
		where = append(where, "municipio_uid = ?")
		args = append(args, filter.MunicipioUID)
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

func (r *sqliteRepo) FindColoniaByCodigoID(codigoID string) (*domain.Colonia, error) {
	query := "SELECT " + coloniaViewSelect + " FROM vw_colonias_busqueda v WHERE v.codigo_id = ? LIMIT 1"
	row := r.db.QueryRow(query, codigoID)

	colonia, err := scanColonia(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &colonia, nil
}

func (r *sqliteRepo) FindNearestColoniasByCodigoID(codigoID string, limit int) ([]domain.ColoniaCercana, error) {
	query := `
		WITH origen AS (
			SELECT codigo_id, centro_lat AS lat, centro_lon AS lon
			FROM vw_colonias_busqueda
			WHERE codigo_id = ?
			  AND centro_lat IS NOT NULL
			  AND centro_lon IS NOT NULL
		)
		SELECT ` + coloniaViewSelect + `,
		       ((v.centro_lat - o.lat) * (v.centro_lat - o.lat) + (v.centro_lon - o.lon) * (v.centro_lon - o.lon)) AS dist2
		FROM vw_colonias_busqueda v
		JOIN origen o
		WHERE v.codigo_id <> o.codigo_id
		  AND v.centro_lat IS NOT NULL
		  AND v.centro_lon IS NOT NULL
		ORDER BY dist2 ASC, v.codigo, v.colonia_nombre
		LIMIT ?`

	f, err := r.db.Query(query, codigoID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var resultados []domain.ColoniaCercana
	for f.Next() {
		row, scanErr := scanColoniaCercana(f)
		if scanErr != nil {
			return nil, scanErr
		}
		resultados = append(resultados, row)
	}
	if err := f.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

func (r *sqliteRepo) FindNearestColoniasByCP(cp string, limit int) ([]domain.ColoniaCercana, error) {
	query := `
		WITH origen AS (
			SELECT codigo_id, centro_lat AS lat, centro_lon AS lon
			FROM vw_colonias_busqueda
			WHERE codigo = ?
			  AND centro_lat IS NOT NULL
			  AND centro_lon IS NOT NULL
		),
		distancias AS (
			SELECT v.codigo_id,
			       MIN((v.centro_lat - o.lat) * (v.centro_lat - o.lat) + (v.centro_lon - o.lon) * (v.centro_lon - o.lon)) AS dist2
			FROM vw_colonias_busqueda v
			JOIN origen o
			WHERE v.centro_lat IS NOT NULL
			  AND v.centro_lon IS NOT NULL
			  AND NOT EXISTS (SELECT 1 FROM origen x WHERE x.codigo_id = v.codigo_id)
			GROUP BY v.codigo_id
			ORDER BY dist2 ASC
			LIMIT ?
		)
		SELECT ` + coloniaViewSelect + `,
		       d.dist2
		FROM distancias d
		JOIN vw_colonias_busqueda v ON v.codigo_id = d.codigo_id
		ORDER BY d.dist2 ASC, v.codigo, v.colonia_nombre`

	f, err := r.db.Query(query, cp, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var resultados []domain.ColoniaCercana
	for f.Next() {
		row, scanErr := scanColoniaCercana(f)
		if scanErr != nil {
			return nil, scanErr
		}
		resultados = append(resultados, row)
	}
	if err := f.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}
