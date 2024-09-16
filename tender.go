package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Tender struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"` 
	CreatorUsername *string   `json:"creatorid"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}


func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	var tender Tender
	if err := json.NewDecoder(r.Body).Decode(&tender); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userExists bool
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM employee WHERE username = $1)`, tender.CreatorUsername).Scan(&userExists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !userExists {
		http.Error(w, "Invalid creator username", http.StatusBadRequest)
		return
	}

	var id string
	err = db.QueryRow(`INSERT INTO tenders (name, description, status, creator_id, version, created_at, updated_at)
		VALUES ($1, $2, $3, (SELECT id FROM employee WHERE username = $4), 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id`,
		tender.Name, tender.Description, tender.Status, tender.CreatorUsername,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tender.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
	json.NewEncoder(w).Encode(tender)
}


func GetTenderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var tender Tender
	err := db.QueryRow(`SELECT id, name, description, status, creator_id, version, created_at, updated_at FROM tenders WHERE id = $1`, id).Scan(
		&tender.ID, &tender.Name, &tender.Description, &tender.Status, &tender.CreatorUsername, &tender.Version, &tender.CreatedAt, &tender.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
	json.NewEncoder(w).Encode(tender)
}

func PublishTenderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := db.Exec(`UPDATE tenders SET status = 'PUBLISHED', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func CancelTenderHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := db.Exec(`UPDATE tenders SET status = 'CANCELED', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var tender Tender
	if err := json.NewDecoder(r.Body).Decode(&tender); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
		UPDATE tenders
		SET name = $1, description = $2, status = $3, version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4`,
		tender.Name, tender.Description, tender.Status, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tender.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
	json.NewEncoder(w).Encode(tender)
}

func GetAllTendersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, description, status, creator_id, version, created_at, updated_at FROM tenders")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tenders []Tender
	for rows.Next() {
		var t Tender
		var creatorUsername sql.NullString 

		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &creatorUsername, &t.Version, &t.CreatedAt, &t.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if creatorUsername.Valid {
			t.CreatorUsername = &creatorUsername.String
		} else {
			t.CreatorUsername = nil
		}

		tenders = append(tenders, t)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
	json.NewEncoder(w).Encode(tenders)
}
