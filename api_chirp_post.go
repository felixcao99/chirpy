package main

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/felixcao99/chirpy/internal/database"
	"github.com/google/uuid"
)

func postChirpsHandler(w http.ResponseWriter, r *http.Request) {
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
