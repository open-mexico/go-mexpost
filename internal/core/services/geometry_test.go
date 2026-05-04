package services_test

import (
	"testing"

	"github.com/open-mexico/go-mexpost/internal/core/services"
	"github.com/stretchr/testify/assert"
)

func TestPointInGeoJSON_InsideSquare(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}`

	inside, err := services.PointInGeoJSON(5, 5, geo)
	assert.NoError(t, err)
	assert.True(t, inside)
}

func TestPointInGeoJSON_OutsideSquare(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}`

	inside, err := services.PointInGeoJSON(15, 5, geo)
	assert.NoError(t, err)
	assert.False(t, inside)
}

func TestPointInGeoJSON_BoundarySquare(t *testing.T) {
	geo := `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}`

	inside, err := services.PointInGeoJSON(0, 5, geo)
	assert.NoError(t, err)
	assert.True(t, inside)
}
