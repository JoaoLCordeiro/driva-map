package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/empresa/mercado/internal/handlers"
	"github.com/empresa/mercado/internal/middleware"
	"github.com/empresa/mercado/internal/repository"
	_ "github.com/lib/pq"
)

func main() {
	// connecting to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("db open:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}
	log.Println("database connected")

	repo := repository.New(db)
	h := handlers.New(repo)

	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("/auth/login", h.Login)

	// states
	mux.HandleFunc("/states", h.ListStates)
	mux.HandleFunc("/states/", h.GetState)

	// branches
	mux.HandleFunc("/branches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			middleware.Auth(h.CreateBranch)(w, r)
			return
		}
		h.ListBranches(w, r)
	})
	mux.HandleFunc("/branches/", func(w http.ResponseWriter, r *http.Request) {
		// reject sub-paths like /branches/1/something
		path := strings.TrimPrefix(r.URL.Path, "/branches/")
		if strings.Contains(path, "/") {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodPut:
			middleware.Auth(h.UpdateBranch)(w, r)
		case http.MethodDelete:
			middleware.Auth(h.DeleteBranch)(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", middleware.CORS(mux)))
}