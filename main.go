package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"

	"github.com/felixcao99/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

var apiCfg *apiConfig

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	dbQueries := database.New(db)

	apiCfg = &apiConfig{}
	apiCfg.dbQueries = dbQueries
	apiCfg.platform = platform

	serverMux := http.NewServeMux()
	serverMux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serverMux.HandleFunc("GET /api/healthz", healthzHandler)
	// serverMux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	// serverMux.HandleFunc("GET /api/metrics", metricsHandler)
	serverMux.HandleFunc("POST /api/users", userHandler)
	serverMux.HandleFunc("GET /admin/metrics", adminMetricsHandler)
	// serverMux.HandleFunc("POST /api/reset", metricsReset)
	serverMux.HandleFunc("POST /admin/reset", metricsReset)
	serverMux.HandleFunc("POST /api/chirps", chirpsHandler)
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

// func metricsHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)
// 	res := fmt.Sprintf("Hits: %d\n", apiCfg.fileserverHits.Load())
// 	w.Write([]byte(res))
// }

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
	if apiCfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden"))
		return
	}
	err := apiCfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	apiCfg.fileserverHits.Store(0)
}

// func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
// 	type chirpRequest struct {
// 		Chirp string `json:"body"`
// 	}
// 	// type validResponse struct {
// 	// 	Valid bool `json:"valid"`
// 	// }
// 	type errorResponse struct {
// 		Error string `json:"error"`
// 	}
// 	type cleanedResponse struct {
// 		Cleanedbody string `json:"cleaned_body"`
// 	}

// 	var replaced string

// 	filter := []string{"kerfuffle", "sharbert", "fornax"}

// 	decoder := json.NewDecoder(r.Body)
// 	chirpbody := chirpRequest{}
// 	err := decoder.Decode(&chirpbody)
// 	if err == nil {
// 		if len(chirpbody.Chirp) <= 140 {
// 			replaced = chirpbody.Chirp
// 			for _, badword := range filter {
// 				re := regexp.MustCompile("(?i)" + badword)
// 				replaced = re.ReplaceAllString(replaced, "****")
// 			}
// 			validres := cleanedResponse{Cleanedbody: replaced}
// 			validjson, _ := json.Marshal(validres)
// 			w.WriteHeader(http.StatusOK)
// 			w.Header().Set("Content-Type", "application/json")
// 			w.Write(validjson)
// 			return
// 		} else {
// 			errdres := errorResponse{Error: "Chirp is too long"}
// 			errson, _ := json.Marshal(errdres)
// 			w.WriteHeader(400)
// 			w.Header().Set("Content-Type", "application/json")
// 			w.Write(errson)
// 			return
// 		}
// 	} else {
// 		errdres := errorResponse{Error: "Invalid JSON"}
// 		errson, _ := json.Marshal(errdres)
// 		w.WriteHeader(400)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write(errson)
// 		return
// 	}
// }

func userHandler(w http.ResponseWriter, r *http.Request) {
	type userEmail struct {
		Email string `json:"email"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	type userResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	useremail := userEmail{}
	err := decoder.Decode(&useremail)
	if err != nil {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	user, err := apiCfg.dbQueries.CreateUser(r.Context(), useremail.Email)
	if err != nil {
		errdres := errorResponse{Error: "Database error"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	res := userResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Chirp  string `json:"body"`
		UserID string `json:"user_id"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	type chirpResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Chirp     string `json:"body"`
		UserID    string `json:"user_id"`
	}

	var replaced string
	filter := []string{"kerfuffle", "sharbert", "fornax"}
	var createPara database.CreateChirpParams

	decoder := json.NewDecoder(r.Body)
	chirpbody := chirpRequest{}
	err := decoder.Decode(&chirpbody)
	if err == nil {
		if len(chirpbody.Chirp) <= 140 {
			replaced = chirpbody.Chirp
			for _, badword := range filter {
				re := regexp.MustCompile("(?i)" + badword)
				replaced = re.ReplaceAllString(replaced, "****")
			}
			createPara.Body = replaced
			createPara.UserID, _ = uuid.Parse(chirpbody.UserID)

			chirp, err := apiCfg.dbQueries.CreateChirp(r.Context(), createPara)
			if err != nil {
				errdres := errorResponse{Error: "Database error"}
				errson, _ := json.Marshal(errdres)
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "application/json")
				w.Write(errson)
				return
			}
			chirpres := chirpResponse{
				Id:        chirp.ID.String(),
				CreatedAt: chirp.CreatedAt.String(),
				UpdatedAt: chirp.UpdatedAt.String(),
				Chirp:     chirp.Body,
				UserID:    chirp.UserID.String(),
			}

			validjson, _ := json.Marshal(chirpres)
			w.WriteHeader(201)
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
