# Referencia de Endpoints de la API

Base URL: `http://localhost:8080`

---

## GET /colonias

Busca colonias con múltiples combinaciones de filtros.

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `cp` | string | condicional* | Código postal, entre 3 y 5 dígitos (ej. `067`, `0671`, `06700`) |
| `nombre` | string | condicional* | Nombre de colonia exacto o parcial (mínimo 3 chars, case-insensitive) |
| `municipio_id` | string | no | Filtra por ID de municipio |
| `municipio_uid` | string | no | Filtra por UID único de municipio (ej. `09-015`) |
| `incluir_municipio` | bool | no (default: `false`) | Si es `true`, agrega el nombre del municipio validado por `municipio_id` + `estado_id` |
| `incluir_geo` | bool | no (default: `false`) | Si es `true`, incluye el polígono GeoJSON y el centroide |
| `solo_geo` | bool | no (default: `false`) | Si es `true`, devuelve solo colonias que tienen geometría registrada |
| `limit` | int | no | Máximo de resultados a devolver (default: `100`, máx: `500`; con `incluir_geo=true` default: `50`, máx: `100`) |
| `pagina` | int | no (default: `1`) | Número de página (empieza en `1`). Usa junto con `limit` para navegar resultados. |

\* Al menos `cp` o `nombre` es obligatorio.

### Combinaciones soportadas

| # | Parámetros | Descripción |
|---|---|---|
| 1 | `?cp=06700` | CP exacto |
| 2 | `?cp=067` | CP por prefijo parcial |
| 3 | `?nombre=Roma` | Nombre parcial (case-insensitive) |
| 4 | `?cp=067&nombre=Roma` | CP + nombre combinados |
| 5 | `?nombre=Centro&municipio_id=014` | Nombre + municipio (resuelve duplicados) |
| 6 | `?cp=067&municipio_id=015` | CP + municipio |
| 7 | `?solo_geo=true&cp=067` | Solo colonias con polígono GeoJSON |
| 8 | `?cp=067&incluir_geo=true` | Incluye geometría en la respuesta |
| 9 | `?cp=067&incluir_municipio=true` | Incluye el nombre del municipio en la respuesta |
| 10 | `?cp=067&municipio_uid=09-015` | CP + UID único de municipio |

### Respuesta exitosa (200 OK)

```json
{
  "resultados": [
    {
      "codigo_id": "09-015-06700-ROMA NORTE",
      "codigo": "06700",
      "nombre": "Roma Norte",
      "tipo": "Colonia",
      "ciudad": "Ciudad de México",
      "zona": "Urbano",
      "estado_id": "09",
      "municipio_id": "015",
      "municipio_uid": "09-015",
      "municipio_nombre": "Cuauhtémoc"
    }
  ],
  "total": 42,
  "limit": 10,
  "pagina": 1,
  "total_paginas": 5,
  "pagina_anterior": null,
  "pagina_siguiente": "/colonias?cp=067&limit=10&pagina=2"
}
```

Con `incluir_geo=true`:

```json
{
  "resultados": [
    {
      "codigo_id": "09-015-06700-ROMA NORTE",
      "codigo": "06700",
      "nombre": "Roma Norte",
      "tipo": "Colonia",
      "ciudad": "Ciudad de México",
      "zona": "Urbano",
      "estado_id": "09",
      "municipio_id": "015",
      "municipio_uid": "09-015",
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
| `400 Bad Request` | Falta `cp` y `nombre` | `{"error": "parametros invalidos", "detalle": "debes proporcionar cp o nombre"}` |
| `400 Bad Request` | `cp` con menos de 3 o más de 5 caracteres | `{"error": "parametros invalidos", "detalle": "cp invalido: usa entre 3 y 5 digitos (ej. 067 o 06700)"}` |
| `400 Bad Request` | `cp` contiene caracteres no numéricos | `{"error": "parametros invalidos", "detalle": "cp invalido: solo se permiten digitos"}` |
| `400 Bad Request` | Valor inválido en `incluir_geo` o `solo_geo` | `{"error": "incluir_geo debe ser true o false"}` |
| `400 Bad Request` | Valor inválido en `incluir_municipio` | `{"error": "incluir_municipio debe ser true o false"}` |
| `400 Bad Request` | Valor inválido en `limit` | `{"error": "limit debe ser un número entero positivo"}` |
| `400 Bad Request` | `pagina` con valor menor a 1 o no numérico | `{"error": "parametros invalidos", "detalle": "pagina debe ser un número entero mayor a 0"}` |
| `404 Not Found` | Ninguna colonia coincide con los filtros | `{"error": "no se encontraron resultados", "detalle": "ajusta tus filtros e intenta de nuevo"}` |
| `500 Internal Server Error` | Error inesperado de base de datos | `{"error": "error interno", "detalle": "ocurrio un error procesando la solicitud"}` |

---

## GET /colonias/id/:codigo_id

Busca una colonia por su identificador único (`codigo_id`).

### Path Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `codigo_id` | string | sí | ID único de colonia generado en ETL |

### Query Parameters

| Parámetro | Tipo | Requerido | Descripción |
|---|---|---|---|
| `incluir_geo` | bool | no (default: `false`) | Si es `true`, incluye `geometria`, `centro_lon` y `centro_lat` |
| `incluir_municipio` | bool | no (default: `true`) | Si es `true`, incluye `municipio_nombre` |

### Respuesta exitosa (200 OK)

```json
{
  "resultado": {
    "codigo_id": "09-015-06700-ROMA NORTE",
    "codigo": "06700",
    "nombre": "Roma Norte",
    "tipo": "Colonia",
    "ciudad": "Ciudad de México",
    "zona": "Urbano",
    "estado_id": "09",
    "municipio_id": "015",
    "municipio_uid": "09-015",
    "municipio_nombre": "Cuauhtémoc"
  }
}
```

### Respuestas de error

| Código | Causa |
|---|---|
| `400 Bad Request` | `incluir_geo` o `incluir_municipio` inválidos |
| `400 Bad Request` | `codigo_id` vacío |
| `404 Not Found` | No existe colonia para ese `codigo_id` |

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
	  "estado_id": "14",
	  "municipio_uid": "14-120"
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
    "zona": "Urbano",
    "estado_id": "09",
    "municipio_id": "015"
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
    "zona": "Urbano",
    "estado_id": "09",
    "municipio_id": "015",
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

### Los campos `centro_lon` / `centro_lat`

Representan el centroide geográfico de la colonia, útil para centrar un mapa. Igual que `geometria`, solo se incluyen cuando `incluir_geo=true`.

### El campo `municipio_nombre`

Se incluye solo cuando `incluir_municipio=true`. Para evitar ambigüedades, se resuelve con `municipio_id` y `estado_id` (no solo por `municipio_id`).

### Los campos `codigo_id` y `municipio_uid`

`codigo_id` identifica de forma única cada colonia en la base nueva. `municipio_uid` identifica de forma única al municipio y puede usarse para filtrar resultados en `/colonias`.
