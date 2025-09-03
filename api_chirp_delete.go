package main

import (
	"encoding/json"
	"net/http"

	"github.com/felixcao99/chirpy/internal/auth"

	"github.com/google/uuid"
)

func deleteChirpByIDHandler(w http.ResponseWriter, r *http.Request) {

	type errorResponse struct {
		Error string `json:"error"`
	}

	type successResponse struct {
		Message string `json:"message"`
	}

	jwttoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	userid, err := auth.ValidateJWT(jwttoken, apiCfg.jwtscecret)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
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

	chirpUserID := chirp.UserID

	if chirpUserID != userid {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	err = apiCfg.dbQueries.DeleteChirpByID(r.Context(), chirpuuid)
	if err != nil {
		errdres := errorResponse{Error: "Database error"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	deletechirpres := successResponse{
		Message: "Chirp deleted",
	}
	resjson, _ := json.Marshal(deletechirpres)
	w.WriteHeader(204)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}
