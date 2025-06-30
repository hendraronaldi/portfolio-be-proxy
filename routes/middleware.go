package routes

import (
	"log"
	"net/http"
	"portfolio-be-proxy/config"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/cors"
)

const (
	rateLimitPeriod      = 1 * time.Minute
	rateLimitMaxRequests = 8
)

var (
	mu            sync.Mutex
	requestCounts int
	lastResetTime time.Time
)

func init() {
	lastResetTime = time.Now()
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

// --- Logging Middleware ---
// LoggingMiddleware logs details about each incoming request and its corresponding response.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Record the start time of the request

		// Wrap the ResponseWriter to capture the status code and size.
		// We initialize status to 200 OK, as that's the default if WriteHeader isn't explicitly called.
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Log incoming request details BEFORE passing to the next handler
		log.Printf("Incoming Request: %s %s from %s", r.Method, r.URL.Path, getClientIP(r))

		// Call the next handler in the chain, passing our wrapped recorder
		next.ServeHTTP(recorder, r)

		// Log response details AFTER the handler has finished executing
		duration := time.Since(start)
		log.Printf("Response Sent: %s %s - Status: %d, Size: %d bytes, Duration: %s",
			r.Method, r.URL.Path, recorder.status, recorder.size, duration)
	})
}

func ConfigureCors(router http.Handler, frontendOrigin string) http.Handler {
	// CORS options: Define what origins, methods, and headers are allowed.
	c := cors.New(cors.Options{
		// AllowedOrigins:
		// CRUCIAL: Replace with the EXACT URL(s) of your frontend application.
		// For local dev: "http://localhost:5173"
		// For deployed: "https://your-firebase-project.web.app" or "https://www.your-custom-domain.com"
		AllowedOrigins: []string{frontendOrigin},

		// AllowedMethods:
		// Explicitly list the HTTP methods your frontend will use.
		// OPTIONS is essential for CORS preflight requests.
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Common methods, adjust as needed. POST and OPTIONS are vital for your proxy.

		// AllowedHeaders:
		// CRUCIAL: List all non-standard headers your frontend sends.
		// "Content-Type" is often needed for JSON POST requests.
		// "X-API-Key" is your custom header, so it MUST be included here.
		AllowedHeaders: []string{"Content-Type", "X-API-Key"},

		// ExposedHeaders:
		// (Optional) Headers that the browser is allowed to read from the response.
		// ExposedHeaders: []string{"Link"},

		// AllowCredentials:
		// Set to true if your frontend sends credentials like cookies, HTTP authentication,
		// or client certificates. If true, AllowedOrigins cannot contain "*" and must be
		// specific origins.
		AllowCredentials: true,

		// MaxAge:
		// (Optional) How long (in seconds) the results of a preflight request
		// can be cached. Reduces preflight overhead.
		MaxAge: 300, // 5 minutes

		// Debug:
		// Set to true to print detailed CORS debugging information to your server logs.
		// VERY useful during development. Set to false for production.
		Debug: true,
	})

	// The .Handler() method wraps your provided router (e.g., your http.NewServeMux)
	// and returns an http.Handler that applies the CORS rules.
	return c.Handler(router)
}

func globalRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if time.Since(lastResetTime) > rateLimitPeriod {
			requestCounts = 0
			lastResetTime = time.Now()
		}

		requestCounts++

		if requestCounts > rateLimitMaxRequests {
			log.Println("Global Rate limit exceeded")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			log.Println("APIKeyMiddleware: Skipping API key check for OPTIONS preflight request.")
			next.ServeHTTP(w, r)
			return
		}

		if config.Config.RequiredAPIKey == "" {
			log.Printf("API key validation is skipped. Request from %s", getClientIP(r))
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			log.Printf("Unauthorized access attempt from %s: X-API-Key header missing", getClientIP(r))
			http.Error(w, "Unauthorized: X-API-Key header missing", http.StatusUnauthorized)
			return
		}
		if apiKey != config.Config.RequiredAPIKey {
			log.Printf("Forbidden access attempt from %s: Invalid API Key", getClientIP(r))
			http.Error(w, "Forbidden: Invalid API Key", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xRealIP := r.Header.Get("X-Real-Ip"); xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
