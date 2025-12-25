package transport

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig configures CORS behavior for HTTP transports.
type CORSConfig struct {
	// AllowOrigins is a list of origins that are allowed.
	// Use "*" to allow all origins, or specify exact origins.
	AllowOrigins []string

	// AllowMethods is a list of allowed HTTP methods.
	// Default: GET, POST, OPTIONS
	AllowMethods []string

	// AllowHeaders is a list of allowed request headers.
	// Default: Content-Type, Authorization, X-Request-ID
	AllowHeaders []string

	// ExposeHeaders is a list of headers the browser is allowed to access.
	ExposeHeaders []string

	// AllowCredentials indicates whether credentials are allowed.
	AllowCredentials bool

	// MaxAge indicates how long preflight results can be cached (in seconds).
	// Default: 86400 (24 hours)
	MaxAge int
}

// DefaultCORSConfig returns a permissive CORS configuration suitable for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		MaxAge:       86400,
	}
}

// CORSHandler wraps an http.Handler with CORS support.
func CORSHandler(config CORSConfig, next http.Handler) http.Handler {
	// Set defaults
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400
	}

	allowAllOrigins := len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*"
	allowedOrigins := make(map[string]bool)
	for _, origin := range config.AllowOrigins {
		allowedOrigins[origin] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		var allowOrigin string
		if allowAllOrigins {
			allowOrigin = "*"
		} else if origin != "" && allowedOrigins[origin] {
			allowOrigin = origin
		}

		// Set CORS headers if origin is allowed
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Set expose headers for actual requests
			if len(config.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}
		}

		next.ServeHTTP(w, r)
	})
}

// WithCORS configures CORS for the HTTP transport.
func WithCORS(config CORSConfig) HTTPOption {
	return func(h *HTTP) {
		h.corsConfig = &config
	}
}

// WithDefaultCORS enables CORS with default permissive settings.
func WithDefaultCORS() HTTPOption {
	config := DefaultCORSConfig()
	return func(h *HTTP) {
		h.corsConfig = &config
	}
}
