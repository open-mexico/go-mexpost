package handler

import (
	"errors"
	"fmt"
	"math"
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
	CodigoID        string   `json:"codigo_id,omitempty"`
	Codigo          string   `json:"codigo"`
	Nombre          string   `json:"nombre"`
	Tipo            string   `json:"tipo"`
	Ciudad          string   `json:"ciudad"`
	Zona            string   `json:"zona"`
	EstadoID        string   `json:"estado_id"`
	MunicipioID     string   `json:"municipio_id"`
	MunicipioUID    string   `json:"municipio_uid,omitempty"`
	MunicipioNombre *string  `json:"municipio_nombre,omitempty"`
	Geometria       *string  `json:"geometria,omitempty"`
	CentroLon       *float64 `json:"centro_lon,omitempty"`
	CentroLat       *float64 `json:"centro_lat,omitempty"`
}

type municipioResponse struct {
	ID           string `json:"id"`
	Nombre       string `json:"nombre"`
	EstadoID     string `json:"estado_id"`
	MunicipioUID string `json:"municipio_uid,omitempty"`
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

	incluirMunicipio, ok := parseBoolQuery(c, "incluir_municipio", false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incluir_municipio debe ser true o false"})
		return
	}

	limit, ok := parseIntQuery(c, "limit", 0)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit debe ser un número entero positivo"})
		return
	}

	pagina, ok := parseIntQuery(c, "pagina", 1)
	if !ok || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pagina debe ser un número entero mayor a 0"})
		return
	}

	// Calcular límite y offset efectivos
	maxLimit := ports.MaxLimit
	defaultLimit := ports.DefaultLimit
	if incluirGeo {
		maxLimit = ports.MaxLimitGeo
		defaultLimit = ports.DefaultLimitGeo
	}
	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}
	offset := (pagina - 1) * limit

	filtro := ports.ColoniaSearchFilter{
		CP:               c.Query("cp"),
		Nombre:           c.Query("nombre"),
		MunicipioID:      c.Query("municipio_id"),
		MunicipioUID:     c.Query("municipio_uid"),
		IncluirMunicipio: incluirMunicipio,
		SoloGeo:          soloGeo,
		Limit:            limit,
		Offset:           offset,
	}

	total, err := h.service.ContarColonias(filtro)
	if err != nil {
		h.writeError(c, err)
		return
	}
	if total == 0 {
		h.writeError(c, domain.ErrNotFound)
		return
	}

	colonias, err := h.service.BuscarColonias(filtro, incluirGeo)
	if err != nil {
		h.writeError(c, err)
		return
	}

	resp := make([]coloniaResponse, 0, len(colonias))
	for _, col := range colonias {
		resp = append(resp, toColoniaResponse(col, incluirGeo, incluirMunicipio))
	}

	totalPaginas := int(math.Ceil(float64(total) / float64(limit)))
	baseURL := fmt.Sprintf("%s?%s", c.Request.URL.Path, buildQueryWithoutPagina(c))

	var paginaAnterior, paginaSiguiente *string
	if pagina > 1 {
		s := fmt.Sprintf("%s&pagina=%d", baseURL, pagina-1)
		paginaAnterior = &s
	}
	if pagina < totalPaginas {
		s := fmt.Sprintf("%s&pagina=%d", baseURL, pagina+1)
		paginaSiguiente = &s
	}

	c.JSON(http.StatusOK, gin.H{
		"resultados":       resp,
		"total":            total,
		"limit":            limit,
		"pagina":           pagina,
		"total_paginas":    totalPaginas,
		"pagina_anterior":  paginaAnterior,
		"pagina_siguiente": paginaSiguiente,
	})
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
		resp = append(resp, municipioResponse{ID: m.ID, Nombre: m.Nombre, EstadoID: m.EstadoID, MunicipioUID: m.MunicipioUID})
	}

	c.JSON(http.StatusOK, gin.H{"resultados": resp})
}

func (h *HttpHandler) BuscarColoniaPorID(c *gin.Context) {
	incluirGeo, ok := parseBoolQuery(c, "incluir_geo", false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incluir_geo debe ser true o false"})
		return
	}

	incluirMunicipio, ok := parseBoolQuery(c, "incluir_municipio", true)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incluir_municipio debe ser true o false"})
		return
	}

	colonia, err := h.service.BuscarColoniaPorID(c.Param("codigo_id"), incluirGeo, incluirMunicipio)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"resultado": toColoniaResponse(*colonia, incluirGeo, incluirMunicipio)})
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

	c.JSON(http.StatusOK, gin.H{"resultado": toColoniaResponse(*colonia, incluirGeo, false)})
}

func (h *HttpHandler) writeError(c *gin.Context, err error) {
	var ve domain.ValidationError
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "parametros invalidos",
			"detalle": ve.Error(),
		})
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"detalle": "ajusta tus filtros e intenta de nuevo",
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "error interno",
		"detalle": "ocurrio un error procesando la solicitud",
	})
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

func parseIntQuery(c *gin.Context, key string, defaultValue int) (int, bool) {
	v := c.Query(key)
	if v == "" {
		return defaultValue, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// buildQueryWithoutPagina devuelve el query string actual sin el parámetro pagina,
// para que las URLs de paginación puedan añadirlo de forma limpia.
func buildQueryWithoutPagina(c *gin.Context) string {
	q := c.Request.URL.Query()
	q.Del("pagina")
	return q.Encode()
}

func toColoniaResponse(col domain.Colonia, incluirGeo bool, incluirMunicipio bool) coloniaResponse {
	resp := coloniaResponse{
		CodigoID:     col.CodigoID,
		Codigo:       col.Codigo,
		Nombre:       col.Nombre,
		Tipo:         col.Tipo,
		Ciudad:       col.Ciudad,
		Zona:         col.Zona,
		EstadoID:     col.EstadoID,
		MunicipioID:  col.MunicipioID,
		MunicipioUID: col.MunicipioUID,
	}

	if incluirMunicipio {
		resp.MunicipioNombre = col.MunicipioNombre
	}

	if incluirGeo {
		resp.Geometria = col.Geometria
		resp.CentroLon = col.CentroLon
		resp.CentroLat = col.CentroLat
	}

	return resp
}
