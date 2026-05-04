package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

const epsilon = 1e-9

type geoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// PointInGeoJSON valida si un punto lon/lat cae dentro de una geometria GeoJSON.
// Soporta Polygon y MultiPolygon.
func PointInGeoJSON(lon, lat float64, geoJSONString string) (bool, error) {
	var geo geoJSONGeometry
	if err := json.Unmarshal([]byte(geoJSONString), &geo); err != nil {
		return false, fmt.Errorf("geojson invalido: %w", err)
	}

	switch geo.Type {
	case "Polygon":
		var polygon [][][]float64
		if err := json.Unmarshal(geo.Coordinates, &polygon); err != nil {
			return false, fmt.Errorf("coordenadas de poligono invalidas: %w", err)
		}
		return pointInPolygonWithHoles(lon, lat, polygon)
	case "MultiPolygon":
		var multi [][][][]float64
		if err := json.Unmarshal(geo.Coordinates, &multi); err != nil {
			return false, fmt.Errorf("coordenadas de multipoligono invalidas: %w", err)
		}
		for _, polygon := range multi {
			inside, err := pointInPolygonWithHoles(lon, lat, polygon)
			if err != nil {
				return false, err
			}
			if inside {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, errors.New("tipo de geometria no soportado")
	}
}

func pointInPolygonWithHoles(lon, lat float64, rings [][][]float64) (bool, error) {
	if len(rings) == 0 {
		return false, errors.New("poligono sin anillos")
	}

	insideOuter, err := pointInRing(lon, lat, rings[0])
	if err != nil || !insideOuter {
		return false, err
	}

	for i := 1; i < len(rings); i++ {
		insideHole, err := pointInRing(lon, lat, rings[i])
		if err != nil {
			return false, err
		}
		if insideHole {
			return false, nil
		}
	}

	return true, nil
}

func pointInRing(x, y float64, ring [][]float64) (bool, error) {
	if len(ring) < 4 {
		return false, errors.New("anillo invalido")
	}

	inside := false
	j := len(ring) - 1
	for i := 0; i < len(ring); i++ {
		if len(ring[i]) < 2 || len(ring[j]) < 2 {
			return false, errors.New("punto de anillo invalido")
		}

		xi, yi := ring[i][0], ring[i][1]
		xj, yj := ring[j][0], ring[j][1]

		if pointOnSegment(x, y, xi, yi, xj, yj) {
			return true, nil
		}

		intersects := (yi > y) != (yj > y)
		if intersects {
			xCross := (xj-xi)*(y-yi)/(yj-yi) + xi
			if x < xCross {
				inside = !inside
			}
		}
		j = i
	}

	return inside, nil
}

func pointOnSegment(px, py, x1, y1, x2, y2 float64) bool {
	sqLen := (x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)
	if sqLen <= epsilon {
		return math.Abs(px-x1) <= epsilon && math.Abs(py-y1) <= epsilon
	}

	cross := (px-x1)*(y2-y1) - (py-y1)*(x2-x1)
	if math.Abs(cross) > epsilon {
		return false
	}

	dot := (px-x1)*(x2-x1) + (py-y1)*(y2-y1)
	if dot < -epsilon {
		return false
	}
	return dot <= sqLen+epsilon
}
