package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
)

const filepathRoot = "."
const port = "8080"
const chirpCharLimit = 140

type apiConfig struct { 
  fileserverHits int
}
type body struct {
  Body string `json:"body"`
}
type errorReply struct {
  Error string `json:"error"`
}
type validReply struct {
  Valid bool `json:"valid"`
}
type cleanReply struct {
  CleanedBody string `json:"cleaned_body"`
}
func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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

func jsonReplies(w http.ResponseWriter, jsonReply errorReply, status int) {
    eJsonReply, err := json.Marshal(jsonReply)
    if err != nil {
      log.Printf("Error marshalling JSON: %s", err)
      w.WriteHeader(500)
      return
    }
    w.WriteHeader(status)
    w.Write(eJsonReply)
}

func replaceProfane(bod *body) {
    newBody := bod.Body
    profanities := [3]string{"kerfuffle", "sharbert", "fornax"}
    
    for _, pw := range profanities {
        re := regexp.MustCompile(`(?i)\b` + pw + `\b`)
        newBody = re.ReplaceAllString(newBody, "****")
    }
    
    bod.Body = newBody
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  
  decoder := json.NewDecoder(r.Body)
  bod := body{}
  decodingErr := decoder.Decode(&bod)

  if decodingErr != nil {
    jsonReply := errorReply{
      Error: "Something went wrong",
    }
    log.Printf(jsonReply.Error)

    jsonReplies(w, jsonReply, 500)
    return
  }

  if (len(bod.Body) > chirpCharLimit) {
    jsonReply := errorReply{
      Error: "Chirp is too long",
    }
    log.Printf(jsonReply.Error)

    jsonReplies(w, jsonReply, 400)
    return
  }

  replaceProfane(&bod)
  clean := cleanReply {
    CleanedBody: bod.Body,
  }

  validJsonReply, err := json.Marshal(clean)
  if err != nil {
    log.Printf("Error marshalling JSON: %s", err)
    w.WriteHeader(500)
  }
  // response 200 is implied
  w.Write(validJsonReply)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
  mux := http.NewServeMux() 

  apiCfg := apiConfig {
    fileserverHits: 0,
  }

  // or http.Dir("./app")
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)
  mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)

  // wrap the mux to add CORS 
  corsMux := middlewareCors(mux)

  // don't use http.HandleFunc; instead create instance and assign the handler like this:
  server := &http.Server {
    Addr:    ":" + port,
    Handler: corsMux,
  }

  log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
  // log.Fatal(http.ListenAndServe(":8080", nil)) becomes:
  log.Fatal(server.ListenAndServe())
}
