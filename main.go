package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	var err error
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func main() {
	InitDB()
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/api/ping", PingHandler).Methods("GET")
	r.HandleFunc("/api/tenders/list", GetAllTendersHandler).Methods("GET")
	r.HandleFunc("/api/tenders/new", CreateTenderHandler).Methods("POST")
	r.HandleFunc("/api/tenders/{id}/publish", PublishTenderHandler).Methods("POST")
	r.HandleFunc("/api/tenders/{id}/cancel", CancelTenderHandler).Methods("POST")

	r.HandleFunc("/api/tenders/{id}", GetTenderHandler).Methods("GET")
	r.HandleFunc("/api/tenders/{id}/edit", EditTenderHandler).Methods("PATCH")
	r.HandleFunc("/api/bids/new", CreateBidHandler).Methods("POST")
    r.HandleFunc("/api/bids/{id}/publish", PublishBidHandler).Methods("POST")
    r.HandleFunc("/api/bids/{id}/cancel", CancelBidHandler).Methods("POST")
    r.HandleFunc("/api/bids/{id}/edit", EditBidHandler).Methods("PUT")
    r.HandleFunc("/api/bids/{id}/status", UpdateBidStatusHandler).Methods("PATCH")


	log.Fatal(http.ListenAndServe(":8080", r))
}