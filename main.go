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

func main() {
  mux := http.NewServeMux()

  mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

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
