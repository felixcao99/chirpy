package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func allChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type chirpResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Chirp     string `json:"body"`
		UserID    string `json:"user_id"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	chirps, err := apiCfg.dbQueries.AllChirps(r.Context())
	if err != nil {
		errdres := errorResponse{Error: "Database error"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	var res []chirpResponse
	for _, chirp := range chirps {
		chirpres := chirpResponse{
			Id:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
			Chirp:     chirp.Body,
			UserID:    chirp.UserID.String(),
		}
		res = append(res, chirpres)
	}

	resjson, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	type chirpResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Chirp     string `json:"body"`
		UserID    string `json:"user_id"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	chirpID := r.PathValue("chirpID")
	chirpuuid, err := uuid.Parse(chirpID)
	if err != nil {
		errdres := errorResponse{Error: "Invalid chirp ID"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirpByID(r.Context(), chirpuuid)
	if err != nil {
		errdres := errorResponse{Error: "Chirp not found"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(404)
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

	resjson, _ := json.Marshal(chirpres)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}
