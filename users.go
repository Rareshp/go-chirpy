package main

import (
	"encoding/json"
	"net/http"
  "sort"
  "strconv"
  "errors"
)

type User struct {
  Hash string `json:"hash"`
  Email string `json:"email"`
  ID int `json:"id"`
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  RefreshTokenRevokedAt string `json:"refresh_token_revoked_at"`
  AccessTokenRevokedAt string `json:"access_token_revoked_at"`
}

type UserResponse struct {
  Email string `json:"email"`
  ID int `json:"id"`
}

func (cfg *apiConfig) findUserByEmail(email string) (User, error) {
	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
    return User{}, err
	}
  
	for _, dbUser := range dbUsers {
    return dbUser, nil
	}

  return User{}, nil
}

func (cfg *apiConfig) findUserById(id int) (User, error) {
	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		return User{}, err
	}

  // equivalent to saying the id exceeds available chirps
  if len(dbUsers) < id {
		return User{}, errors.New("user not found")
  }

  user := User{
    ID: dbUsers[id-1].ID,
    Email: dbUsers[id-1].Email,
  }

  return user, nil
}

func (cfg *apiConfig) validateUserEmail(email string) (error) {
  user, err := cfg.findUserByEmail(email)
	if err != nil {
    return err
	}
  
  if user.Email != "" {
    return errors.New("this email is already used")
	}

  return nil
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
    Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

  // check if email is in use
	emailErr := cfg.validateUserEmail(params.Email)
	if emailErr != nil {
		respondWithError(w, http.StatusBadRequest, emailErr.Error())
		return
	}

	user, err := cfg.DB.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, UserResponse{
		Email: user.Email,
		ID:   user.ID,
	})
}

func (cfg *apiConfig) handlerUsersRetrieve(w http.ResponseWriter, r *http.Request) {
	dbUsers, err := cfg.DB.GetUsers()
	if err == nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't login. Email not in use")
		return
	}

	users := []User{}
	for _, dbUser := range dbUsers {
		users = append(users, User{
			ID:   dbUser.ID,
			Email: dbUser.Email,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	respondWithJSON(w, http.StatusOK, users)
}

func (cfg *apiConfig) handlerUsersRetrieveById(w http.ResponseWriter, r *http.Request) {
  // get the id
  id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve id from the GET request")
		return
	}

  user, err := cfg.findUserById(id) 
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve id from the GET request")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
    Password string `json:"password"`
	}
  type UserResponse struct {
    Email string `json:"email"`
    ID int `json:"id"`
    Token string `json:"token"`
    Refresh_Token string `json:"refresh_token"`
  }

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

  // check if email is in use
	user, emailErr := cfg.findUserByEmail(params.Email)
	if emailErr != nil {
		respondWithError(w, http.StatusBadRequest, emailErr.Error())
		return
	}

  // check user password 
  passErr := CheckPasswordHash(params.Password, user.Hash)
	if passErr != nil {
		respondWithError(w, http.StatusUnauthorized, passErr.Error())
		return
	}

  accessToken, TokenErr := cfg.jwtCreateAccessToken(user.ID)
  if TokenErr != nil {
		respondWithError(w, http.StatusInternalServerError, TokenErr.Error())
    return
  }
  refreshToken, TokenErr := cfg.jwtCreateRefreshToken(user.ID)
  if TokenErr != nil {
		respondWithError(w, http.StatusInternalServerError, TokenErr.Error())
    return
  }

  cfg.DB.SetUserTokens(user.ID, accessToken, refreshToken)

	respondWithJSON(w, http.StatusOK, UserResponse{
		Email: user.Email,
		ID:   user.ID,
    Token: accessToken,
    Refresh_Token: refreshToken,
	})
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
  type UserResponse struct {
    Email string `json:"email"`
    ID int `json:"id"`
  }

	token, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	subject, err := ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

  issuer, errI := GetIssuer(token, cfg.jwtSecret)
  if errI != nil {
    respondWithError(w, http.StatusUnauthorized, "Coulnd't validate Issuer")
    return
  }
  if issuer == "chirpy-refresh" {
    respondWithError(w, http.StatusUnauthorized, "Cannot use refresh token for this entry point")
    return
  }

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	user, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusOK, UserResponse{
    ID:    user.ID,
    Email: user.Email,
	})
}
