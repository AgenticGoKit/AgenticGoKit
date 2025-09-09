package webui

import (
	"net/http"
	"strings"
	"time"
)

// withMiddleware wraps the HTTP handler with all necessary middleware
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied is executed first)
	handler := next
	handler = s.withLogging(handler)
	handler = s.withErrorRecovery(handler)
	handler = s.withCORS(handler)
	handler = s.withSecurityHeaders(handler)

	return handler
}

// withCORS adds CORS headers to handle cross-origin requests
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from localhost during development
		origin := r.Header.Get("Origin")
		if origin == "" || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// In production, you might want to be more restrictive
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// withSecurityHeaders adds basic security headers
func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - basic policy for now
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:"
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}

// withLogging adds request logging
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(wrapper, r)

		// Log the request
		duration := time.Since(start)
		s.logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status_code", wrapper.statusCode).
			Dur("duration", duration).
			Msg("HTTP request")
	})
}

// withErrorRecovery recovers from panics and handles errors gracefully
func (s *Server) withErrorRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error().
					Interface("error", err).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("Panic recovered in HTTP handler")

				// Return a generic error response
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write ensures status code is captured even if WriteHeader is not called explicitly
func (rw *responseWriter) Write(b []byte) (int, error) {
	// If WriteHeader hasn't been called, it will be called with StatusOK
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// RateLimitConfig configures rate limiting middleware
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	Enabled           bool
}

// withRateLimit adds rate limiting middleware (disabled by default for Task 1)
func (s *Server) withRateLimit(next http.Handler, config RateLimitConfig) http.Handler {
	if !config.Enabled {
		return next
	}

	// TODO: Implement rate limiting in future tasks
	// For now, just pass through
	return next
}

// RequestMetrics holds metrics for monitoring
type RequestMetrics struct {
	TotalRequests     int64
	ErrorRequests     int64
	AverageResponse   time.Duration
	ActiveConnections int32
}

// getMetrics returns current server metrics (stub for future monitoring)
func (s *Server) getMetrics() RequestMetrics {
	// TODO: Implement proper metrics collection
	return RequestMetrics{
		TotalRequests:     0,
		ErrorRequests:     0,
		AverageResponse:   0,
		ActiveConnections: 0,
	}
}
