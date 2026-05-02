package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/core/ports"
)

type HttpHandler struct {
	service ports.ColoniaService
}

func NewHttpHandler(service ports.ColoniaService) *HttpHandler {
	return &HttpHandler{service: service}
}

func (h *HttpHandler) BuscarColonias(c *gin.Context) {
	cp := c.Query("cp")
	nombre := c.Query("nombre")

	colonias, err := h.service.BuscarColonias(cp, nombre)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(colonias) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No se encontraron resultados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resultados": colonias})
}