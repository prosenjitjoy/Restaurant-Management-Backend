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

var menuCollection = database.OpenCollection(db, "menus")

func GetMenus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR menu IN menus LIMIT 10 RETURN menu"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		menus := []model.Menu{}
		for {
			var menu model.Menu
			_, err := cursor.ReadDocument(context.TODO(), &menu)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read menu items"})
				return
			}

			menus = append(menus, menu)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(menus)
	}
}

func GetMenuByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		menuID := chi.URLParam(r, "menu_id")
		var menu model.Menu

		_, err := menuCollection.ReadDocument(context.TODO(), menuID, &menu)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch menu item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(menu)
	}
}

func CreateMenu() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var menu model.Menu
		err := json.NewDecoder(r.Body).Decode(&menu)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		err = validate.Struct(menu)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
			return
		}

		menu.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.MenuID = uuid.NewString()

		meta, err := menuCollection.CreateDocument(context.TODO(), menu)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create menu item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func UpdateMenuByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		menuID := chi.URLParam(r, "menu_id")
		var menu model.Menu
		err := json.NewDecoder(r.Body).Decode(&menu)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "kindly retype tye the time"})
				return
			}

			updateObject["start_date"] = menu.StartDate
			updateObject["end_date"] = menu.EndDate
		}
		if menu.Name != "" {
			updateObject["name"] = menu.Name
		}
		if menu.Category != "" {
			updateObject["category"] = menu.Category
		}

		menu.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = menu.UpdatedAt

		meta, err := menuCollection.UpdateDocument(context.TODO(), menuID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create menu item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteMenuByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		menuID := chi.URLParam(r, "menu_id")

		meta, err := menuCollection.RemoveDocument(context.TODO(), menuID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete menu item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(check) && end.After(start)
}
