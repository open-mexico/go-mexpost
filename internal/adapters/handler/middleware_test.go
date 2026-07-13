package handler_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/open-mexico/go-mexpost/internal/adapters/handler"
	"github.com/stretchr/testify/assert"
)

func TestSlogMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	router := gin.New()
	router.Use(handler.SlogMiddleware(logger))

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.GET("/error", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})

	t.Run("Logs success request", func(t *testing.T) {
		buf.Reset()
		req, _ := http.NewRequest("GET", "/test?foo=bar", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		logOutput := buf.String()
		assert.Contains(t, logOutput, "\"msg\":\"HTTP request\"")
		assert.Contains(t, logOutput, "\"method\":\"GET\"")
		assert.Contains(t, logOutput, "\"path\":\"/test\"")
		assert.Contains(t, logOutput, "\"status\":200")
		assert.Contains(t, logOutput, "\"query\":\"foo=bar\"")
	})

	t.Run("Logs error request", func(t *testing.T) {
		buf.Reset()
		req, _ := http.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		logOutput := buf.String()
		assert.Contains(t, logOutput, "\"msg\":\"Server error\"")
		assert.Contains(t, logOutput, "\"status\":500")
	})
}
