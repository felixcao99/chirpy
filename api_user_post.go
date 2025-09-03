package main

import (
	"encoding/json"
	"net/http"

	"github.com/felixcao99/chirpy/internal/auth"
	"github.com/felixcao99/chirpy/internal/database"
	"github.com/google/uuid"

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
		Red       bool   `json:"is_chirpy_red"`
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

	var isChirpyRed bool
	if user.IsChirpyRed.Valid {
		isChirpyRed = user.IsChirpyRed.Bool
	}

	res := userResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		Red:       isChirpyRed,
	}

	resjson, _ := json.Marshal(res)
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Password  string `json:"password"`
		Email     string `json:"email"`
		Expiresin int    `json:"expires_in_seconds"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type loginResponse struct {
		Id         string `json:"id"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
		Email      string `json:"email"`
		Token      string `json:"token"`
		FreshToken string `json:"refresh_token"`
		Red        bool   `json:"is_chirpy_red"`
	}

	var insertfreshtoken database.InsertFreshTokenParams

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

	// if loginrequest.Expiresin > 0 && loginrequest.Expiresin < 3600 {
	// 	expireseconds = time.Duration(loginrequest.Expiresin) * time.Second
	// } else {
	// 	expireseconds = time.Duration(3600) * time.Second
	// }

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

	// jwttoken, err := auth.MakeJWT(user.ID, apiCfg.jwtscecret, expireseconds)
	refreshtoken, _ := auth.MakeRefreshToken()
	insertfreshtoken.Token = refreshtoken
	insertfreshtoken.UserID = user.ID
	insertedfreshtoken, err := apiCfg.dbQueries.InsertFreshToken(r.Context(), insertfreshtoken)
	if err != nil {
		errdres := errorResponse{Error: "Fresh token not generated"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	jwttoken, err := auth.MakeJWT(user.ID, apiCfg.jwtscecret)
	if err != nil {
		errdres := errorResponse{Error: "Token not generated"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	var isChirpyRed bool
	if user.IsChirpyRed.Valid {
		isChirpyRed = user.IsChirpyRed.Bool
	}

	res := loginResponse{
		Id:         user.ID.String(),
		CreatedAt:  user.CreatedAt.String(),
		UpdatedAt:  user.UpdatedAt.String(),
		Email:      user.Email,
		Token:      jwttoken,
		FreshToken: insertedfreshtoken.Token,
		Red:        isChirpyRed,
	}

	resjson, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type validResponse struct {
		AccessToken string `json:"token"`
	}

	freshtoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	storedfreshtoken, err := apiCfg.dbQueries.GetFreshTokenByToken(r.Context(), freshtoken)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	validuserid, err := auth.ValidateFreshToken(storedfreshtoken)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	accesstoken, err := auth.MakeJWT(validuserid, apiCfg.jwtscecret)
	if err != nil {
		errdres := errorResponse{Error: "Access Token not generated"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	res := validResponse{
		AccessToken: accesstoken,
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func revokeRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type successResponse struct {
		Message string `json:"message"`
	}

	freshtoken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	storedfreshtoken, err := apiCfg.dbQueries.GetFreshTokenByToken(r.Context(), freshtoken)
	if err != nil {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	err = apiCfg.dbQueries.RevokeRefreshToken(r.Context(), storedfreshtoken.Token)
	if err != nil {
		errdres := errorResponse{Error: "Database error"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	res := successResponse{
		Message: "Refresh token revoked for user " + storedfreshtoken.UserID.String(),
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(204)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}

func polkaWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	type polkaData struct {
		UserID string `json:"user_id"`
	}
	type polkaRequest struct {
		Event string    `json:"event"`
		Data  polkaData `json:"data"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type eventResponse struct {
		Event string `json:"event"`
	}

	type userResponse struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
		Red       bool   `json:"is_chirpy_red"`
	}

	apikey, err := auth.GetAPIKey(r.Header)
	if err != nil || apikey != apiCfg.polkakey {
		errdres := errorResponse{Error: "Not Authorized"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	decoder := json.NewDecoder(r.Body)
	polkarequest := polkaRequest{}
	err = decoder.Decode(&polkarequest)
	if err != nil {
		errdres := errorResponse{Error: "Invalid JSON"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}

	if polkarequest.Event != "user.upgraded" {
		eventres := eventResponse{Event: polkarequest.Event}
		evenson, _ := json.Marshal(eventres)
		w.WriteHeader(204)
		w.Header().Set("Content-Type", "application/json")
		w.Write(evenson)
		return
	}

	userid, err := uuid.Parse(polkarequest.Data.UserID)
	if err != nil {
		errdres := errorResponse{Error: "Invalid UserID"}
		errson, _ := json.Marshal(errdres)
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		w.Write(errson)
		return
	}
	user, err := apiCfg.dbQueries.UpdateUserRed(r.Context(), userid)
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
		Red:       user.IsChirpyRed.Bool,
	}
	resjson, _ := json.Marshal(res)
	w.WriteHeader(204)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resjson)
}
