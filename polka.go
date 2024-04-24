
package main 

import (
  "net/http"
	"encoding/json"
)

func (cfg *apiConfig) handlerUserUpgradeToRed(w http.ResponseWriter, r *http.Request) {
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
