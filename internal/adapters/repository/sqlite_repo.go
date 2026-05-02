package repository

import (
	"database/sql"
	"fmt"

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

func (r *sqliteRepo) Buscar(filtroTipo string, valor string) ([]domain.Colonia, error) {
	// Preparamos la consulta permitiendo búsquedas parciales (LIKE valor%)
	columna := "codigo"
	if filtroTipo == "nombre" {
		columna = "nombre"
	}

	query := fmt.Sprintf("SELECT codigo, nombre, tipo, ciudad, zona, estado_id, municipio_id, geometria FROM colonias WHERE %s LIKE ?", columna)
	parametro := valor + "%" // Agregamos el comodín de SQLite al final

	filas, err := r.db.Query(query, parametro)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	var resultados []domain.Colonia
	for filas.Next() {
		var col domain.Colonia
		err := filas.Scan(&col.Codigo, &col.Nombre, &col.Tipo, &col.Ciudad, &col.Zona, &col.EstadoID, &col.MunicipioID, &col.Geometria)
		if err != nil {
			return nil, err
		}
		resultados = append(resultados, col)
	}
	return resultados, nil
}