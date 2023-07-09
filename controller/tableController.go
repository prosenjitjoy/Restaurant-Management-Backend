package controller

import (
	"context"
	"encoding/json"
	"main/database"
	"main/model"
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var tableCollection = database.OpenCollection(db, "tables")

func GetTables() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR table IN tables LIMIT 10 RETURN table"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		tables := []model.Table{}
		for {
			var table model.Table
			_, err := cursor.ReadDocument(context.TODO(), &table)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read table items"})
				return
			}

			tables = append(tables, table)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tables)
	}
}

func GetTableByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableID := chi.URLParam(r, "table_id")
		var table model.Table

		_, err := tableCollection.ReadDocument(context.TODO(), tableID, &table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch table item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(table)
	}
}

func CreateTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var table model.Table
		err := json.NewDecoder(r.Body).Decode(&table)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		err = validate.Struct(table)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
			return
		}

		table.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.TableID = uuid.NewString()

		meta, err := tableCollection.CreateDocument(context.TODO(), table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create table item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func UpdateTableByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableID := chi.URLParam(r, "table_id")
		var table model.Table
		err := json.NewDecoder(r.Body).Decode(&table)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if table.NumberOfGuest != nil {
			updateObject["number_of_guest"] = table.NumberOfGuest
		}
		if table.TableNumber != nil {
			updateObject["table_number"] = table.NumberOfGuest
		}

		table.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = table.UpdatedAt

		meta, err := tableCollection.UpdateDocument(context.TODO(), tableID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to update table item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteTableByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableID := chi.URLParam(r, "table_id")

		meta, err := tableCollection.RemoveDocument(context.TODO(), tableID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete table item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}
