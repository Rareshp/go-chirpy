package main

import (
	"log"
	"net/http"
  "os"
  "github.com/joho/godotenv"
)

const filepathRoot = "."
const port = "8080"
const chirpCharLimit = 140
const dbPath = "database.json"

type apiConfig struct { 
  fileserverHits int
  DB             *DB
  jwtSecret       string
  polkaAPIKey     string
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

func healthHandler(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
  // read from .env
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
  jwtSecret := os.Getenv("JWT_SECRET")
  polkaAPIKey := os.Getenv("POLKA_API_KEY")

  mux := http.NewServeMux() 


	db, err := NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

  apiCfg := apiConfig {
    fileserverHits: 0,
    DB: db,
    jwtSecret: jwtSecret,
    polkaAPIKey: polkaAPIKey,
  }

  // or http.Dir("./app")
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("GET /api/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)

	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.handlerChirpsRetrieveById)
	mux.HandleFunc("DELETE /api/chirps/{id}", apiCfg.handlerChirpsDeleteById)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUserCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)
	mux.HandleFunc("GET /api/users/{id}", apiCfg.handlerUsersRetrieveById)

	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)
  mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
  mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeToken)

  mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUserUpgradeToRed)

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
