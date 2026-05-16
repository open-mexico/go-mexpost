package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu      sync.Mutex
	clients map[string]*ipLimiter
	r       rate.Limit
	burst   int
}

func newRateLimiterStore(r rate.Limit, burst int) *rateLimiterStore {
	store := &rateLimiterStore{
		clients: make(map[string]*ipLimiter),
		r:       r,
		burst:   burst,
	}
	// Limpieza periódica: elimina IPs que no han sido vistas en los últimos 5 minutos.
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.mu.Lock()
			for ip, cl := range store.clients {
				if time.Since(cl.lastSeen) > 5*time.Minute {
					delete(store.clients, ip)
				}
			}
			store.mu.Unlock()
		}
	}()
	return store
}

func (s *rateLimiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	cl, ok := s.clients[ip]
	if !ok {
		cl = &ipLimiter{limiter: rate.NewLimiter(s.r, s.burst)}
		s.clients[ip] = cl
	}
	cl.lastSeen = time.Now()
	return cl.limiter
}

// RateLimitMiddleware limita a `requestsPerSecond` solicitudes por segundo por IP,
// con un burst máximo de `burst` solicitudes. Los valores recomendados para una
// API pública son r=10, burst=20 (10 req/s sostenidos, ráfagas de hasta 20).
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	store := newRateLimiterStore(rate.Limit(requestsPerSecond), burst)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !store.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "demasiadas solicitudes",
				"detalle": "excediste el límite de solicitudes, intenta de nuevo en unos segundos",
			})
			return
		}
		c.Next()
	}
}
