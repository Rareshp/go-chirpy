package main

import (
	"encoding/json"
	"net/http"
  "errors"
  "strings"
  "sort"
  "strconv"
)

type Chirp struct {
  Body string `json:"body"`
  ID int `json:"id"`
  Author_ID int `json:"author_id"`
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	msgCleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	_, errS := ValidateJWT(token, cfg.jwtSecret)
	if errS != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

  issuer, errI := GetIssuer(token, cfg.jwtSecret)
  if errI != nil {
    respondWithError(w, http.StatusUnauthorized, "Coulnd't validate Issuer")
    return
  }
  if issuer != "chirpy-access" {
    respondWithError(w, http.StatusUnauthorized, "Must use access token for this entry point")
    return
  }

  user, errU := cfg.DB.FindUserByAccessToken(token)
  if errU != nil {
    respondWithError(w, http.StatusUnauthorized, "Cannoy find user with this token")
  }

	chirp, err := cfg.DB.CreateChirp(msgCleaned, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		Body: chirp.Body,
		ID:   chirp.ID,
    Author_ID: user.ID,
	})
}
func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
      Author_ID: dbChirp.Author_ID,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsRetrieveById(w http.ResponseWriter, r *http.Request) {
  // get the id
  id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve id from the GET request")
		return
	}

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

  // equivalent to saying the id exceeds available chirps
  if len(dbChirps) < id {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
    return
  }

  chirp := Chirp{
    ID: dbChirps[id-1].ID,
    Body: dbChirps[id-1].Body,
  }

	respondWithJSON(w, http.StatusOK, chirp)
}
