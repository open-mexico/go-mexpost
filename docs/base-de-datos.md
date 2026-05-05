# Base de Datos

## Archivo `mapa.db`

Go-MexPost utiliza una base de datos **SQLite embebida** (`mapa.db`) que se descarga desde el repositorio ETL oficial ([sepomex-db-generator](https://github.com/open-mexico/sepomex-db-generator)). No requiere instalación de ningún servidor de base de datos.

El driver utilizado es [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite), una implementación 100% en Go (sin CGO) que garantiza compatibilidad cross-platform y despliegue simplificado.

---

## Tablas

### `colonias` — 145,420 registros

Tabla principal del sistema. Contiene todos los códigos postales de México con sus datos textuales y geoespaciales.

| Columna | Tipo | Descripción |
|---|---|---|
| `codigo` | TEXT | Código postal (5 dígitos, p.ej. `06700`) |
| `nombre` | TEXT | Nombre de la colonia (p.ej. `Roma Norte`) |
| `tipo` | TEXT | Tipo de asentamiento (p.ej. `Colonia`, `Fraccionamiento`, `Pueblo`) |
| `ciudad` | TEXT | Ciudad o delegación |
| `zona` | TEXT | Tipo de zona (`Urbana`, `Rural`, `Semiurbana`) |
| `estado_id` | TEXT | ID del estado (2 chars, p.ej. `09`) |
| `municipio_id` | TEXT | ID del municipio (3 chars, p.ej. `014`) |
| `geometria` | TEXT | Polígono GeoJSON como string (puede ser NULL) |
| `min_lon` | REAL | Longitud mínima del Bounding Box |
| `min_lat` | REAL | Latitud mínima del Bounding Box |
| `max_lon` | REAL | Longitud máxima del Bounding Box |
| `max_lat` | REAL | Latitud máxima del Bounding Box |
| `centro_lon` | REAL | Longitud del centroide (puede ser NULL) |
| `centro_lat` | REAL | Latitud del centroide (puede ser NULL) |

#### ¿Qué es el Bounding Box?

El BBox es el rectángulo mínimo que encierra completamente el polígono de una colonia. Se usa como filtro rápido en la geocodificación inversa antes de aplicar Ray-Casting:

```
        max_lat
           │
   ┌───────┴───────┐
   │   polígono    │ max_lon
   │    real       │ ─────
   └───────┬───────┘
           │
        min_lat
   min_lon
```

---

### `municipios` — 2,458 registros

Catálogo de municipios de México.

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | TEXT | Identificador del municipio (3 chars, p.ej. `014`) |
| `nombre` | TEXT | Nombre del municipio (p.ej. `Benito Juárez`) |
| `estado_id` | TEXT | ID del estado al que pertenece |

---

### `estados` — 32 registros

Catálogo de los 32 estados de la República Mexicana.

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | TEXT | Identificador del estado (2 chars, p.ej. `09`) |
| `nombre` | TEXT | Nombre del estado (p.ej. `Ciudad de México`) |

---

## Índices

La base de datos cuenta con **15 índices** optimizados para las consultas más frecuentes:

### Índices en `colonias`

| Índice | Columnas | Propósito |
|---|---|---|
| `idx_colonia_codigo` | `codigo` | Búsqueda por CP exacto o prefijo |
| `idx_colonia_nombre` | `nombre COLLATE NOCASE` | Búsqueda por nombre (case-insensitive) |
| `idx_colonia_codigo_nombre` | `(codigo, nombre)` | Búsqueda combinada CP + nombre |
| `idx_colonia_estado_nombre` | `(estado_id, nombre)` | Filtro por estado + nombre |
| `idx_colonia_estado_codigo` | `(estado_id, codigo)` | Filtro por estado + CP |
| `idx_colonia_municipio_nombre` | `(municipio_id, nombre)` | Filtro por municipio + nombre |
| `idx_colonia_municipio_codigo` | `(municipio_id, codigo)` | Filtro por municipio + CP |
| `idx_colonia_geo_not_null` | `codigo WHERE geometria IS NOT NULL` | Índice parcial para colonias con geometría |
| `idx_colonia_bbox` | `(min_lat, max_lat, min_lon, max_lon)` | Prefiltrado BBox para geocodificación inversa |
| `idx_colonia_estado_bbox` | `(estado_id, min_lat, max_lat, min_lon, max_lon)` | BBox acotado por estado |
| `idx_colonia_municipio_bbox` | `(municipio_id, min_lat, max_lat, min_lon, max_lon)` | BBox acotado por municipio |
| `idx_colonia_centro` | `(centro_lat, centro_lon)` | Búsquedas por centroide |

### Índices en `municipios`

| Índice | Columnas | Propósito |
|---|---|---|
| `idx_municipio_estado` | `estado_id` | Listar municipios de un estado |
| `idx_municipio_nombre` | `nombre COLLATE NOCASE` | Búsqueda por nombre |

### Índices en `estados`

| Índice | Columnas | Propósito |
|---|---|---|
| `idx_estado_nombre` | `nombre COLLATE NOCASE` | Búsqueda por nombre |

---

## Queries dinámicas del repositorio

El repositorio en `internal/adapters/repository/sqlite_repo.go` construye queries de forma dinámica con parámetros posicionales para prevenir SQL injection.

### `SearchColonias` — combinaciones de filtros

```sql
-- Solo por CP (parcial ≥3 chars)
SELECT codigo, nombre, ... FROM colonias WHERE codigo LIKE '067%' ORDER BY codigo, nombre

-- Solo por nombre (parcial ≥3 chars, case-insensitive)
SELECT codigo, nombre, ... FROM colonias WHERE nombre LIKE '%roma%' COLLATE NOCASE ORDER BY codigo, nombre

-- CP + nombre combinados
SELECT codigo, nombre, ... FROM colonias
  WHERE codigo LIKE '067%'
    AND nombre LIKE '%roma%' COLLATE NOCASE
  ORDER BY codigo, nombre

-- Con municipio_id
SELECT codigo, nombre, ... FROM colonias
  WHERE nombre LIKE '%centro%' COLLATE NOCASE
    AND municipio_id = '014'
  ORDER BY codigo, nombre

-- Solo con geometría (solo_geo=true)
SELECT codigo, nombre, ... FROM colonias
  WHERE codigo LIKE '067%'
    AND geometria IS NOT NULL AND TRIM(geometria) <> ''
  ORDER BY codigo, nombre
```

### `SearchMunicipios`

```sql
-- Por nombre (parcial)
SELECT id, nombre, estado_id FROM municipios WHERE nombre LIKE '%zapopan%' COLLATE NOCASE ORDER BY nombre

-- Por estado
SELECT id, nombre, estado_id FROM municipios WHERE estado_id = '14' ORDER BY nombre
```
