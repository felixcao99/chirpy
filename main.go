package main

import (
	"encoding/json"
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
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
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

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Chirp string `json:"body"`
	}
	type validResponse struct {
		Valid bool `json:"valid"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpbody := chirpRequest{}
	err := decoder.Decode(&chirpbody)
	if err == nil {
		if len(chirpbody.Chirp) <= 140 {
			validres := validResponse{Valid: true}
			validjson, _ := json.Marshal(validres)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write(validjson)
			return
		} else {
			errdres := errorResponse{Error: "Chirp is too long"}
			errson, _ := json.Marshal(errdres)
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			w.Write(errson)
			return
		}
	} else {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
}
