package main

import (
	"encoding/json"
	"net/http"

	"github.com/felixcao99/chirpy/internal/auth"
	"github.com/felixcao99/chirpy/internal/database"

	_ "github.com/lib/pq"
)

func userHandler(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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
	userrequest := userRequest{}
	err := decoder.Decode(&userrequest)
	if err != nil {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	var createPara database.CreateUserParams
	createPara.Email = userrequest.Email
	hashedPassword, _ := auth.HashPassword(userrequest.Password)
	createPara.HashedPassword = hashedPassword

	user, err := apiCfg.dbQueries.CreateUser(r.Context(), createPara)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	type loginResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	loginrequest := loginRequest{}
	err := decoder.Decode(&loginrequest)
	if err != nil {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	user, err := apiCfg.dbQueries.GetUserByEmail(r.Context(), loginrequest.Email)
	if err != nil {
		errdres := errorResponse{Error: "Invalid email or password"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	err = auth.CheckPasswordHash(loginrequest.Password, user.HashedPassword)
	if err != nil {
		errdres := errorResponse{Error: "Invalid email or password"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	res := loginResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}
