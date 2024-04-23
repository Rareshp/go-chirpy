package main 

import (
  "fmt"
  "net/http"
	"encoding/json"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
  type response struct {
    Token string `json:"token"`
  }

	type parameters struct {
		Email string `json:"email"`
    Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if (params.Email != "") || (params.Password != "") {
		respondWithError(w, http.StatusUnauthorized, "This entrypoint does not consume a body")
		return
	}

	token, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	_, errV := ValidateJWT(token, cfg.jwtSecret)
	if errV != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}
  issuer, err := GetIssuer(token, cfg.jwtSecret)
  if err != nil {
    respondWithError(w, http.StatusUnauthorized, "Coulnd't validate Issuer")
    return 
  }
  if issuer != "chirpy-refresh" {
    respondWithError(w, http.StatusUnauthorized, "You must use a refreh token for this entry point")
    return
  }

  user, err := cfg.DB.FindUserByRefreshToken(token)
  if err != nil {
    respondWithError(w, http.StatusUnauthorized, err.Error())
    return
  }
  if user.RefreshTokenRevokedAt != "" {
    respondWithError(w, http.StatusUnauthorized, "This refresh token was revoked")
    return
  }

  // create new token
  // it took me a while to figure out why the tests were failing; 
  // I misunderstood that this "refresh" endpoint should return an access token 
  newAccessToken, TokenErr := cfg.jwtCreateAccessToken(user.ID)
  if TokenErr != nil {
		respondWithError(w, http.StatusInternalServerError, TokenErr.Error())
    return
  }

  cfg.DB.SetUserTokens(user.ID, user.AccessToken, newAccessToken)
  respondWithJSON(w, http.StatusOK, response{
    Token: newAccessToken,
  })

}
func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
  type response struct {
    Token string `json:"token"`
  }
	token, err := GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	_, errV := ValidateJWT(token, cfg.jwtSecret)
	if errV != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

  _, errR := cfg.DB.RevokeRefreshToken(token)
  if errR != nil {
    respondWithError(w, http.StatusInternalServerError, "Coulnd't find token to revoke")
  }

  respondWithJSON(w, http.StatusOK, "")
}
