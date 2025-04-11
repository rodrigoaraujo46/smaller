package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func getSmaller(url string, db *sql.DB) (string, error) {
	smaller := uuid.New().String()[:8]

	query := `INSERT INTO smaller (id, url) VALUES ($1, $2)`
	_, err := db.Exec(query, smaller, url)
	if err != nil {
		return "", fmt.Errorf("error inserting URL into database: %v", err)
	}

	return smaller, nil
}

func retriveURL(uuid string, db *sql.DB) (string, error) {
	fmt.Println(uuid)
	var url string

	query := "SELECT url FROM smaller WHERE id = $1"
	err := db.QueryRow(query, uuid).Scan(&url)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", err
	}

	return url, nil
}

func handle(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "web/index.html")

	case http.MethodPost:
		url := r.FormValue("url")
		shortened, err := getSmaller(url, db)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		url, err = retriveURL(shortened, db)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(shortened + "\n" + url))

	}
}

func main() {
	connStr := "user=smalleruser dbname=smallerurl host=localhost sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handle(w, r, db)
	})

	http.ListenAndServe(":8080", nil)
}
