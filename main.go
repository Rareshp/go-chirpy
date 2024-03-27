package main

import (
  "log"
  "net/http"
)

const filepathRoot = "."
const port = "8080"

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
  mux := http.NewServeMux()

  // or http.Dir("./app")
  mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot))))
  mux.HandleFunc("/healthz", healthHandler)

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
