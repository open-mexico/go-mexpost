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

---

## Implementación en Go

La función `PointInGeoJSON` en `internal/core/services/geometry.go`:

1. Parsea el GeoJSON del string almacenado en la base de datos.
2. Extrae el array de coordenadas del polígono.
3. Itera sobre cada arista del polígono aplicando el test de cruce de Ray-Casting.
4. Devuelve `true` si el número de cruces es impar.

> **Convención GeoJSON**: el primer argumento es **longitud** (X), el segundo **latitud** (Y). Esto sigue la especificación GeoJSON RFC 7946.
