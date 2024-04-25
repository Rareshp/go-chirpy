
package main 

import (
  "net/http"
	"encoding/json"
  "strings"
  "errors"
)

func (cfg *apiConfig) handlerUserUpgradeToRed(w http.ResponseWriter, r *http.Request) {
	apiKey, errK := getAPIKey(r.Header)
	if errK != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
  if apiKey != cfg.polkaAPIKey {
    respondWithError(w, http.StatusUnauthorized, "Invalid API Key")
    return
  }

  type response struct {
    Token string `json:"token"`
  }
  
  type parameterData struct {
    User_Id int `json:"user_id"`
  }

	type parameters struct {
		Event string `json:"event"`
    Data parameterData `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
  if params.Event != "user.upgraded" {
    respondWithJSON(w, http.StatusOK, "")
    return 
  }
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  ok := cfg.DB.UpgradeUserToRed(params.Data.User_Id)
  if ok != nil {
    respondWithError(w, http.StatusInternalServerError, ok.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, "")
}

func getAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("auth header is empty")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
