# Referencia de Endpoints de la API

Base URL: `http://localhost:8080`

---

## GET /colonias

Busca colonias con múltiples combinaciones de filtros.

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `cp` | string | condicional* | Código postal exacto o prefijo (mínimo 3 chars para búsqueda parcial) |
| `nombre` | string | condicional* | Nombre de colonia exacto o parcial (mínimo 3 chars, case-insensitive) |
| `municipio_id` | string | no | Filtra por ID de municipio |
| `incluir_geo` | bool | no (default: `false`) | Si es `true`, incluye el polígono GeoJSON y el centroide |
| `solo_geo` | bool | no (default: `false`) | Si es `true`, devuelve solo colonias que tienen geometría registrada |

\* Al menos `cp` o `nombre` es obligatorio.

### Combinaciones soportadas

| # | Parámetros | Descripción |
|---|---|---|
| 1 | `?cp=06700` | CP exacto |
| 2 | `?cp=067` | CP por prefijo parcial |
| 3 | `?nombre=Roma` | Nombre parcial (case-insensitive) |
| 4 | `?cp=067&nombre=Roma` | CP + nombre combinados |
| 5 | `?nombre=Centro&municipio_id=014` | Nombre + municipio (resuelve duplicados) |
| 6 | `?cp=067&municipio_id=014` | CP + municipio |
| 7 | `?solo_geo=true&cp=067` | Solo colonias con polígono GeoJSON |
| 8 | `?cp=067&incluir_geo=true` | Incluye geometría en la respuesta |

### Respuesta exitosa (200 OK)

```json
{
  "resultados": [
    {
      "codigo": "06700",
      "nombre": "Roma Norte",
      "tipo": "Colonia",
      "ciudad": "Ciudad de México",
      "zona": "Urbana",
      "estado_id": "09",
      "municipio_id": "014"
    }
  ]
}
```

Con `incluir_geo=true`:

```json
{
  "resultados": [
    {
      "codigo": "06700",
      "nombre": "Roma Norte",
      "tipo": "Colonia",
      "ciudad": "Ciudad de México",
      "zona": "Urbana",
      "estado_id": "09",
      "municipio_id": "014",
      "geometria": "{\"type\":\"Polygon\",\"coordinates\":[...]}",
      "centro_lon": -99.1634,
      "centro_lat": 19.4181
    }
  ]
}
```

### Respuestas de error

| Código | Causa | Ejemplo |
|---|---|---|
| `400 Bad Request` | Falta `cp` y `nombre` | `{"error": "debes proporcionar cp o nombre"}` |
| `400 Bad Request` | Valor inválido en `incluir_geo` o `solo_geo` | `{"error": "incluir_geo debe ser true o false"}` |
| `404 Not Found` | Ninguna colonia coincide con los filtros | `{"error": "no se encontraron resultados"}` |
| `500 Internal Server Error` | Error inesperado de base de datos | `{"error": "error interno"}` |

---

## GET /municipios

Busca municipios por nombre y/o estado.

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `nombre` | string | condicional* | Nombre de municipio exacto o parcial (mínimo 3 chars, case-insensitive) |
| `estado_id` | string | condicional* | ID del estado al que pertenece el municipio |

\* Al menos `nombre` o `estado_id` es obligatorio.

### Combinaciones soportadas

| # | Parámetros | Descripción |
|---|---|---|
| 1 | `?nombre=Zapopan` | Municipio por nombre parcial |
| 2 | `?nombre=Centro&estado_id=09` | Nombre filtrado por estado |
| 3 | `?estado_id=14` | Todos los municipios de un estado |

### Respuesta exitosa (200 OK)

```json
{
  "resultados": [
    {
      "id": "120",
      "nombre": "Zapopan",
      "estado_id": "14"
    }
  ]
}
```

### Respuestas de error

| Código | Causa |
|---|---|
| `400 Bad Request` | Falta `nombre` y `estado_id` |
| `404 Not Found` | Ningún municipio coincide |

---

## GET /coordenadas

Geocodificación inversa: dado un punto geográfico, devuelve la colonia a la que pertenece.

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `lat` | float64 | **sí** | Latitud del punto (rango: -90 a 90) |
| `lon` | float64 | **sí** | Longitud del punto (rango: -180 a 180) |
| `estado_id` | string | no | Acota la búsqueda a un estado específico (recomendado para mayor rendimiento) |
| `incluir_geo` | bool | no (default: `false`) | Si es `true`, incluye el polígono GeoJSON y el centroide en la respuesta |

### Ejemplo de uso

```
GET /coordenadas?lat=19.4181&lon=-99.1634
GET /coordenadas?lat=19.4181&lon=-99.1634&estado_id=09
GET /coordenadas?lat=19.4181&lon=-99.1634&incluir_geo=true
```

### Respuesta exitosa (200 OK)

```json
{
  "resultado": {
    "codigo": "06700",
    "nombre": "Roma Norte",
    "tipo": "Colonia",
    "ciudad": "Ciudad de México",
    "zona": "Urbana",
    "estado_id": "09",
    "municipio_id": "014"
  }
}
```

Con `incluir_geo=true`:

```json
{
  "resultado": {
    "codigo": "06700",
    "nombre": "Roma Norte",
    "tipo": "Colonia",
    "ciudad": "Ciudad de México",
    "zona": "Urbana",
    "estado_id": "09",
    "municipio_id": "014",
    "geometria": "{\"type\":\"Polygon\",\"coordinates\":[...]}",
    "centro_lon": -99.1634,
    "centro_lat": 19.4181
  }
}
```

> **Nota**: La respuesta de `/coordenadas` es un objeto único (`"resultado"`), no un array, ya que un punto geográfico pertenece exactamente a una colonia.

### Respuestas de error

| Código | Causa |
|---|---|
| `400 Bad Request` | Falta `lat` o `lon` |
| `400 Bad Request` | `lat` o `lon` no son números válidos |
| `400 Bad Request` | `lat` fuera de rango [-90, 90] o `lon` fuera de [-180, 180] |
| `404 Not Found` | El punto no cae dentro de ninguna colonia registrada |

---

## Notas generales

### Lógica de búsqueda parcial vs. exacta

Para búsquedas de texto (`cp`, `nombre`), la lógica es:

- Si el valor tiene **3 o más caracteres**: se usa `LIKE '%valor%'` (búsqueda parcial, case-insensitive).
- Si el valor tiene **menos de 3 caracteres**: se usa `= valor` (coincidencia exacta).

### El campo `geometria`

El campo `geometria` contiene un GeoJSON válido en texto. Se omite del JSON de respuesta por defecto para reducir el tamaño de las respuestas. Solo se incluye cuando el cliente envía `incluir_geo=true`.

### El campo `centro_lon` / `centro_lat`

Representan el centroide geográfico de la colonia, útil para centrar un mapa. Igual que `geometria`, solo se incluyen cuando `incluir_geo=true`.
