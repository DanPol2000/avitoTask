package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Bid struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	TenderID        string `json:"tenderId"`
	CreatorUsername string `json:"creatorUsername"`
}

func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	var bid Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var tenderExists bool
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM tenders WHERE id = $1)`, bid.TenderID).Scan(&tenderExists)
	if err != nil || !tenderExists {
		http.Error(w, "Invalid tender ID", http.StatusBadRequest)
		return
	}

	var userExists bool
	err = db.QueryRow(`SELECT EXISTS (SELECT 1 FROM employee WHERE username = $1)`, bid.CreatorUsername).Scan(&userExists)
	if err != nil || !userExists {
		http.Error(w, "Invalid creator username", http.StatusBadRequest)
		return
	}

	var id string
	err = db.QueryRow(`
        INSERT INTO bids (name, description, status, tender_id, creator_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, (SELECT id FROM employee WHERE username = $5), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id`,
		bid.Name, bid.Description, bid.Status, bid.TenderID, bid.CreatorUsername,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bid.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
	json.NewEncoder(w).Encode(bid)
}

func UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["id"]
	newStatus := r.URL.Query().Get("status")

	validStatuses := map[string]bool{
		"CREATED":   true,
		"PUBLISHED": true,
		"CANCELED":  true,
	}

	if _, ok := validStatuses[newStatus]; !ok {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`UPDATE bids SET status = $1 WHERE id = $2`, newStatus, bidID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func PublishBidHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["id"]

	result, err := db.Exec(`UPDATE bids SET status = 'PUBLISHED' WHERE id = $1`, bidID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func CancelBidHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["id"]

	result, err := db.Exec(`UPDATE bids SET status = 'CANCELED' WHERE id = $1`, bidID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func EditBidHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["id"]

	var bid Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`UPDATE bids SET name = $1, description = $2 WHERE id = $3`,
		bid.Name, bid.Description, bidID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}
