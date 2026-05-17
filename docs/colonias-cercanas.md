# Funcionamiento tecnico de /colonias/cercanas

Esta guia explica como funciona internamente el endpoint `GET /colonias/cercanas`, para que puedas interpretar mejor los resultados y usarlo correctamente en aplicaciones de mapas, buscadores o sugerencias de zonas vecinas.

## Objetivo del endpoint

Recibir una colonia de referencia y devolver una lista de colonias candidatas ordenadas de la mas cercana a la mas lejana.

La referencia se puede enviar de dos formas:

- `codigo_id`: referencia una colonia unica.
- `cp`: referencia un conjunto de colonias del mismo codigo postal.

## Contrato de entrada

Debes enviar exactamente uno de estos parametros:

- `codigo_id`
- `cp`

Parametros opcionales:

- `limit` (default `20`, maximo `100`)
- `incluir_municipio` (default `true`)
- `incluir_geo` (default `false`)

Validaciones relevantes:

- Si no envias `codigo_id` ni `cp`, responde `400`.
- Si envias ambos, responde `400`.
- Si usas `cp`, debe ser exacto de 5 digitos.

## Como calcula la cercania

La comparacion se hace usando centroides (`centro_lat`, `centro_lon`) de cada colonia en la vista `vw_colonias_busqueda`.

### Caso 1: referencia por `codigo_id`

1. Se obtiene el centroide de la colonia origen.
2. Se comparan todas las demas colonias con centroide valido.
3. Se calcula una distancia en grados al cuadrado.
4. Se ordena ascendente por distancia.

### Caso 2: referencia por `cp`

1. Se toman todas las colonias del codigo postal origen como conjunto de referencia.
2. Para cada colonia candidata, se calcula su distancia contra cada referencia del CP.
3. Se conserva la distancia minima por candidata.
4. Se ordena ascendente por esa minima distancia.

Este enfoque por minima distancia permite resultados mas coherentes cuando un CP contiene varias colonias y esta rodeado por otras.

## Formula de distancia

Internamente se ordena por:

```text
dist2 = (lat2 - lat1)^2 + (lon2 - lon1)^2
```

Y se expone al cliente una distancia aproximada en kilometros:

```text
distancia_km ~= sqrt(dist2) * 111.32
```

Importante:

- `distancia_km` es una aproximacion util para ranking y proximidad relativa.
- No representa distancia por vialidades ni tiempo de traslado.
- Tampoco es distancia borde-a-borde del poligono; es centroide-a-centroide.

## Orden, desempates y exclusion de origen

- El resultado siempre llega ordenado de menor a mayor distancia.
- En empates, se desempata por `codigo` y luego por `colonia_nombre`.
- La colonia de origen no aparece en resultados:
  - Por `codigo_id`, se excluye ese mismo ID.
  - Por `cp`, se excluyen todas las colonias que pertenezcan al conjunto origen.

## Campos de salida

Cada elemento incluye los campos de colonia y adicionalmente:

- `distancia_km`: distancia aproximada hacia la referencia.

Si `incluir_municipio=true`, incluye:

- `municipio_nombre`

Si `incluir_geo=true`, incluye:

- `geometria`
- `centro_lon`
- `centro_lat`

## Recomendaciones de uso

- Usa `codigo_id` cuando tengas una colonia especifica seleccionada por el usuario.
- Usa `cp` cuando el contexto de origen sea una zona postal y no una colonia unica.
- Mantener `limit` entre `10` y `30` suele dar buena experiencia en UI.
- Activa `incluir_geo` solo cuando realmente necesites pintar geometria en mapa.

## Ejemplos

```bash
# Referencia por colonia unica
curl "http://localhost:8080/colonias/cercanas?codigo_id=09-015-06700-ROMA%20NORTE&limit=10"

# Referencia por codigo postal
curl "http://localhost:8080/colonias/cercanas?cp=06700&limit=15"
```

## Errores comunes

- `400`: falta referencia o referencia ambigua (`cp` y `codigo_id` juntos).
- `400`: `cp` invalido para este endpoint (debe ser de 5 digitos).
- `404`: no se encontro referencia valida o no hay colonias candidatas con centroides disponibles.

Para el contrato resumido del endpoint, consulta tambien la referencia principal en [/endpoints](/endpoints).