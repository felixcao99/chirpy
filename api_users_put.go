package main

import (
	"encoding/json"
	"net/http"

	"github.com/felixcao99/chirpy/internal/auth"
	"github.com/felixcao99/chirpy/internal/database"
	// _ "github.com/lib/pq"
)

func userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type successResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	type updateRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var updaterequest database.UpdateUserParams

	accesstoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	userid, err := auth.ValidateJWT(accesstoken, apiCfg.jwtscecret)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	decoder := json.NewDecoder(r.Body)
	updatebody := updateRequest{}
	err = decoder.Decode(&updatebody)
	if err != nil {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	updaterequest.ID = userid
	updaterequest.Email = updatebody.Email
	hashedPassword, _ := auth.HashPassword(updatebody.Password)
	updaterequest.HashedPassword = hashedPassword

	updateduser, err := apiCfg.dbQueries.UpdateUser(r.Context(), updaterequest)
	if err != nil {
		errdres := errorResponse{Error: "Database error"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	res := successResponse{
		Id:        updateduser.ID.String(),
		CreatedAt: updateduser.CreatedAt.String(),
		UpdatedAt: updateduser.UpdatedAt.String(),
		Email:     updateduser.Email,
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}
