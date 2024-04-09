package main

import (
	"encoding/json"
	"net/http"
  "sort"
  "strconv"
)

type User struct {
  Email string `json:"email"`
  ID int `json:"id"`
}

func validateUser(email string) (string, error) {
  // TODO
  return email, nil
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateUser(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := cfg.DB.CreateUser(cleaned)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		Email: user.Email,
		ID:   user.ID,
	})
}
func (cfg *apiConfig) handlerUsersRetrieve(w http.ResponseWriter, r *http.Request) {
	dbUsers, err := cfg.DB.GetUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
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
