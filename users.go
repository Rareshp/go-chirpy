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

func (cfg *apiConfig) validateUserEmail(email string) (error) {
  user, err := cfg.findUserByEmail(email)
	if err != nil {
    return err
	}
  
  if user.Email != "" {
    return errors.New("This email is already used")
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

	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

  // equivalent to saying the id exceeds available chirps
  if len(dbUsers) < id {
		respondWithError(w, http.StatusNotFound, "User not found")
    return
  }

  user := User{
    ID: dbUsers[id-1].ID,
    Email: dbUsers[id-1].Email,
  }

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
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

	respondWithJSON(w, http.StatusOK, UserResponse{
		Email: user.Email,
		ID:   user.ID,
	})
}
