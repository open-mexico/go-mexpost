# Geocodificación Inversa y Algoritmo Ray-Casting

## ¿Qué es la geocodificación inversa?

La geocodificación inversa es el proceso de convertir un par de coordenadas geográficas (latitud, longitud) en una dirección o entidad geográfica con nombre. En este proyecto, dado un punto GPS, la API determina en qué **colonia postal** de México se encuentra ese punto.

---

## Problema que resuelve

Una colonia postal no es un punto en el mapa, sino un **polígono** (área) con forma irregular. Para saber si un punto está "dentro" de una colonia, no basta comparar números; hay que hacer geometría computacional.

Ejemplo:

```
Punto de consulta: lat=19.4181, lon=-99.1634

¿Está dentro de Roma Norte?   ← Necesita álgebra de polígonos
¿Está dentro de Roma Sur?
¿Está dentro de Doctores?
```

---

## Estrategia de dos pasos

Para evitar ejecutar el costoso algoritmo geométrico sobre todas las colonias de México (~100,000), el sistema usa una estrategia de **filtrado progresivo**:

### Paso 1 — Prefiltrado BBox (SQLite, muy rápido)

Cada colonia tiene almacenados cuatro valores numéricos que forman su **Bounding Box** (caja delimitadora mínima):

```
min_lon  min_lat  max_lon  max_lat
  │                          │
  └──── esquina SW ──────────┘
              │
      (rectángulo que encierra el polígono)
```

La query SQLite filtra las colonias cuyo BBox contiene el punto:

```sql
SELECT ...
FROM colonias
WHERE min_lat <= :lat   -- el punto no está por encima del techo
  AND max_lat >= :lat   -- el punto no está por debajo del piso
  AND min_lon <= :lon   -- el punto no está a la derecha
  AND max_lon >= :lon   -- el punto no está a la izquierda
  AND geometria IS NOT NULL
```

Este filtro es extremadamente rápido porque las columnas BBox tienen índices en la base de datos. El resultado es un conjunto pequeño de **colonias candidatas** (generalmente 1–5) cuyas cajas rectangulares contienen el punto.

> Sin este filtro, el algoritmo geométrico tendría que evaluar cada colonia de México, lo que sería inviable en producción.

### Paso 2 — Ray-Casting (Go, en memoria)

Para cada colonia candidata, el servicio aplica el algoritmo **Ray-Casting** (también llamado _Point-in-Polygon_) sobre la geometría GeoJSON real del polígono.

---

## Algoritmo Ray-Casting explicado

El Ray-Casting es el método más común para determinar si un punto está dentro de un polígono. La idea es:

> Lanza un rayo horizontal desde el punto hacia la derecha (dirección +X). Cuenta cuántas veces cruza los bordes del polígono. Si el número de cruces es **impar**, el punto está dentro. Si es **par** (o cero), está fuera.

### Visualización

```
         borde
         del polígono
    ╔════════╗
    ║        ║
    ║  ●─────╫──────→ rayo (1 cruce → DENTRO)
    ║        ║
    ╚════════╝

    ●─────╫──────╫──────→ rayo (2 cruces → FUERA)
```

### Fórmula matemática

Para cada arista del polígono definida por los puntos $(x_i, y_i)$ y $(x_j, y_j)$, se calcula si el rayo horizontal desde $(px, py)$ la cruza:

$$\text{cruza} = (y_i > py) \neq (y_j > py)$$

Si cruza, la coordenada X del punto de cruce es:

$$x_{cruce} = \frac{(x_j - x_i)(py - y_i)}{y_j - y_i} + x_i$$

El punto está "a la izquierda" del cruce si $px < x_{cruce}$, lo que activa el toggle del flag `inside`.

---

## Implementación en Go

El código está en [internal/core/services/geometry.go](../internal/core/services/geometry.go).

### Función pública: `PointInGeoJSON`

```go
func PointInGeoJSON(lon, lat float64, geoJSONString string) (bool, error)
```

- Deserializa el string GeoJSON.
- Soporta dos tipos de geometría:
  - **`Polygon`**: un solo polígono, posiblemente con agujeros (_holes_).
  - **`MultiPolygon`**: múltiples polígonos (colonias con partes separadas). Devuelve `true` si el punto cae en **cualquiera** de ellos.

### Soporte de polígonos con agujeros

Un polígono GeoJSON puede tener anillos interiores (agujeros). La estructura de coordenadas es:

```json
{
  "type": "Polygon",
  "coordinates": [
    [[anillo_exterior]],
    [[agujero_1]],
    [[agujero_2]]
  ]
}
```

La lógica es:
1. El punto debe estar dentro del **anillo exterior** (`rings[0]`).
2. El punto **no** debe estar dentro de ningún agujero (`rings[1..n]`).

```go
func pointInPolygonWithHoles(lon, lat float64, rings [][][]float64) (bool, error) {
    // ¿Está dentro del polígono exterior?
    insideOuter, _ := pointInRing(lon, lat, rings[0])
    if !insideOuter { return false, nil }

    // ¿Está dentro de algún agujero? → entonces está FUERA
    for i := 1; i < len(rings); i++ {
        insideHole, _ := pointInRing(lon, lat, rings[i])
        if insideHole { return false, nil }
    }

    return true, nil
}
```

### Caso especial: punto sobre el borde

Si el punto cae exactamente sobre el borde del polígono, la función `pointOnSegment` lo detecta y devuelve `true` (el punto se considera "dentro").

```go
func pointOnSegment(px, py, x1, y1, x2, y2 float64) bool {
    // Verifica que el cross-product sea ~0 (colineal)
    // y que el punto esté en el rango del segmento
}
```

---

## Flujo completo de `/coordenadas`

```
GET /coordenadas?lat=19.4181&lon=-99.1634&estado_id=09
                │
                ▼
        [ HttpHandler.BuscarCoordenadas ]
        - Parsea lat=19.4181, lon=-99.1634
        - Construye ReverseGeocodeFilter
                │
                ▼
        [ coloniaService.BuscarPorCoordenadas ]
        - Valida lat ∈ [-90,90] ✓
        - Valida lon ∈ [-180,180] ✓
                │
                ▼
        [ sqliteRepo.FindColoniasByPointBBox ]
        - SQL: WHERE min_lat<=19.4181 AND max_lat>=19.4181
                     AND min_lon<=-99.1634 AND max_lon>=-99.1634
                     AND estado_id='09'
        - Devuelve N candidatas (p.ej. Roma Norte, Roma Sur, Doctores)
                │
                ▼ (bucle sobre candidatas)
        [ services.PointInGeoJSON(lon, lat, geometria) ]
        - Deserializa GeoJSON de Roma Norte
        - Ray-Casting: el punto está DENTRO ✓
                │
                ▼
        Devuelve Roma Norte → handler → JSON 200 OK
```

---

## ¿Por qué Ray-Casting y no otro algoritmo?

| Algoritmo | Complejidad | Ventajas | Desventajas |
|---|---|---|---|
| Ray-Casting | O(n) por polígono | Simple, confiable, sin dependencias externas | Requiere manejo especial de bordes |
| Winding Number | O(n) por polígono | Más robusto con bordes | Más complejo de implementar |
| Spatial Index (R-Tree) | O(log n) | Muy rápido a escala | Requiere librería externa, estado en memoria |

Ray-Casting es la elección correcta aquí porque:
- El prefiltrado BBox ya reduce el problema a pocos polígonos.
- No requiere dependencias externas (CGO-free, alineado con el uso de `modernc.org/sqlite`).
- Es determinista y fácil de probar con polígonos simulados.
