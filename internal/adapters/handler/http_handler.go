package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/core/domain"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
)

type HttpHandler struct {
	service ports.ColoniaService
}

func NewHttpHandler(service ports.ColoniaService) *HttpHandler {
	return &HttpHandler{service: service}
}

type coloniaResponse struct {
	Codigo      string   `json:"codigo"`
	Nombre      string   `json:"nombre"`
	Tipo        string   `json:"tipo"`
	Ciudad      string   `json:"ciudad"`
	Zona        string   `json:"zona"`
	EstadoID    string   `json:"estado_id"`
	MunicipioID string   `json:"municipio_id"`
	Geometria   *string  `json:"geometria,omitempty"`
	CentroLon   *float64 `json:"centro_lon,omitempty"`
	CentroLat   *float64 `json:"centro_lat,omitempty"`
}

type municipioResponse struct {
	ID       string `json:"id"`
	Nombre   string `json:"nombre"`
	EstadoID string `json:"estado_id"`
}

func (h *HttpHandler) BuscarColonias(c *gin.Context) {
	incluirGeo, ok := parseBoolQuery(c, "incluir_geo", false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incluir_geo debe ser true o false"})
		return
	}

	soloGeo, ok := parseBoolQuery(c, "solo_geo", false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solo_geo debe ser true o false"})
		return
	}

	filtro := ports.ColoniaSearchFilter{
		CP:          c.Query("cp"),
		Nombre:      c.Query("nombre"),
		MunicipioID: c.Query("municipio_id"),
		SoloGeo:     soloGeo,
	}

	colonias, err := h.service.BuscarColonias(filtro, incluirGeo)
	if err != nil {
		h.writeError(c, err)
		return
	}

	resp := make([]coloniaResponse, 0, len(colonias))
	for _, col := range colonias {
		resp = append(resp, toColoniaResponse(col, incluirGeo))
	}

	c.JSON(http.StatusOK, gin.H{"resultados": resp})
}

func (h *HttpHandler) BuscarMunicipios(c *gin.Context) {
	filtro := ports.MunicipioSearchFilter{
		Nombre:   c.Query("nombre"),
		EstadoID: c.Query("estado_id"),
	}

	municipios, err := h.service.BuscarMunicipios(filtro)
	if err != nil {
		h.writeError(c, err)
		return
	}

	resp := make([]municipioResponse, 0, len(municipios))
	for _, m := range municipios {
		resp = append(resp, municipioResponse{ID: m.ID, Nombre: m.Nombre, EstadoID: m.EstadoID})
	}

	c.JSON(http.StatusOK, gin.H{"resultados": resp})
}

func (h *HttpHandler) BuscarCoordenadas(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat y lon son obligatorios"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat invalida"})
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lon invalida"})
		return
	}

	incluirGeo, ok := parseBoolQuery(c, "incluir_geo", false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incluir_geo debe ser true o false"})
		return
	}

	filtro := ports.ReverseGeocodeFilter{
		Lat:      lat,
		Lon:      lon,
		EstadoID: c.Query("estado_id"),
	}

	colonia, err := h.service.BuscarPorCoordenadas(filtro, incluirGeo)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"resultado": toColoniaResponse(*colonia, incluirGeo)})
}

func (h *HttpHandler) writeError(c *gin.Context, err error) {
	var ve domain.ValidationError
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{"error": ve.Error()})
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
}

func parseBoolQuery(c *gin.Context, key string, defaultValue bool) (bool, bool) {
	v := c.Query(key)
	if v == "" {
		return defaultValue, true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}
	return b, true
}

func toColoniaResponse(col domain.Colonia, incluirGeo bool) coloniaResponse {
	resp := coloniaResponse{
		Codigo:      col.Codigo,
		Nombre:      col.Nombre,
		Tipo:        col.Tipo,
		Ciudad:      col.Ciudad,
		Zona:        col.Zona,
		EstadoID:    col.EstadoID,
		MunicipioID: col.MunicipioID,
	}

	if incluirGeo {
		resp.Geometria = col.Geometria
		resp.CentroLon = col.CentroLon
		resp.CentroLat = col.CentroLat
	}

	return resp
}
