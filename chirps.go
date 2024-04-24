package main

import (
	"encoding/json"
	"net/http"
  "errors"
  "strings"
  "sort"
  "strconv"
)

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

func (cfg *apiConfig) validateToken(r *http.Request, tokenType string) (string, error) {

	token, err := GetBearerToken(r.Header)
	if err != nil {
		return "", err
	}
	_, errS := ValidateJWT(token, cfg.jwtSecret)
	if errS != nil {
		return "", errors.New("could not validate token")
	}

  issuer, errI := GetIssuer(token, cfg.jwtSecret)
  if errI != nil {
    return "", errors.New("could not validate Issuer")
  }
  if issuer != tokenType {
    return "", errors.New("wrong token type")
  }

  return token, nil
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

  token, errToken := cfg.validateToken(r, "chirpy-access")
  if errToken != nil {
    respondWithError(w, http.StatusUnauthorized, errToken.Error())
  }

  user, errU := cfg.DB.FindUserByAccessToken(token)
  if errU != nil {
    respondWithError(w, http.StatusUnauthorized, "Cannot find user with this token")
    return
  }

	chirp, err := cfg.DB.CreateChirp(msgCleaned, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp")
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
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps")
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

func (cfg *apiConfig) retrieveChirpById (w http.ResponseWriter, r *http.Request) (Chirp, error) {
  // get the id
  id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
    return Chirp{}, errors.New("couldn't retrieve id from the GET request")
	}

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
    return Chirp{}, errors.New("couldn't retrieve chirp")
	}

  // equivalent to saying the id exceeds available chirps
  // log.Println(dbChirps)
  if len(dbChirps) < id {
    return Chirp{}, errors.New("chirp not found")
  }

  chirp := Chirp{
    ID: dbChirps[id-1].ID,
    Body: dbChirps[id-1].Body,
    Author_ID: dbChirps[id-1].Author_ID,
  }
  return chirp, nil
}

func (cfg *apiConfig) handlerChirpsRetrieveById(w http.ResponseWriter, r *http.Request) {
  chirp, err := cfg.retrieveChirpById(w, r)
  if err != nil {
    respondWithError(w, http.StatusForbidden, err.Error())
  }
	respondWithJSON(w, http.StatusOK, chirp)
}
func (cfg *apiConfig) handlerChirpsDeleteById(w http.ResponseWriter, r *http.Request) {
  token, errToken := cfg.validateToken(r, "chirpy-access")
  if errToken != nil {
    respondWithError(w, http.StatusForbidden, errToken.Error())
    return
  }

  user, errU := cfg.DB.FindUserByAccessToken(token)
  if errU != nil {
    respondWithError(w, http.StatusForbidden, "Cannot find user with this token")
    return
  }

  chirp, errC := cfg.retrieveChirpById(w, r)
  if errC != nil {
    respondWithError(w, http.StatusForbidden, "Could not find this chrip")
    return
  }

  if chirp.Author_ID != user.ID {
    respondWithError(w, http.StatusForbidden, "This chirp does not belong to you")
    return
  }

  errD := cfg.DB.DeleteChrip(chirp)
  if errD != nil {
    respondWithError(w, http.StatusInternalServerError, errD.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, chirp)
}
