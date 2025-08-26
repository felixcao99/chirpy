package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

var apiCfg *apiConfig

func main() {
	apiCfg = &apiConfig{}
	serverMux := http.NewServeMux()
	serverMux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthzHandler)
	// serverMux.HandleFunc("GET /api/metrics", metricsHandler)
	serverMux.HandleFunc("GET /admin/metrics", adminMetricsHandler)
	// serverMux.HandleFunc("POST /api/reset", metricsReset)
	serverMux.HandleFunc("POST /admin/reset", metricsReset)
	var server http.Server
	server.Addr = ":8080"
	server.Handler = serverMux
	server.ListenAndServe()
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	res := fmt.Sprintf("Hits: %d\n", apiCfg.fileserverHits.Load())
	w.Write([]byte(res))
}

func adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	res := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, apiCfg.fileserverHits.Load())
	w.Write([]byte(res))
}

func metricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	apiCfg.fileserverHits.Store(0)
}
