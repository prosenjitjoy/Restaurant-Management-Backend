package controller

import (
	"context"
	"encoding/json"
	"main/database"
	"main/model"
	"math"
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	validate *validator.Validate = validator.New()
	db       driver.Database     = database.DBinstance()
)

var foodCollection = database.OpenCollection(db, "foods")

type status map[string]interface{}

func GetFoods() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR food IN foods LIMIT 10 RETURN food"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		foods := []model.Food{}
		for {
			var food model.Food
			_, err := cursor.ReadDocument(context.TODO(), &food)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read menu items"})
				return
			}

			foods = append(foods, food)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(foods)
	}
}

func GetFoodByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		foodID := chi.URLParam(r, "food_id")
		var food model.Food

		_, err := foodCollection.ReadDocument(context.TODO(), foodID, &food)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch food item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(food)
	}
}

func CreateFood() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var food model.Food
		err := json.NewDecoder(r.Body).Decode(&food)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		err = validate.Struct(food)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
			return
		}

		var menu model.Menu
		_, err = menuCollection.ReadDocument(context.TODO(), *food.MenuID, &menu)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch menu item"})
			return
		}

		food.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.FoodID = uuid.NewString()
		num := math.Round(*food.UnitPrice*100) / 100
		food.UnitPrice = &num

		meta, err := foodCollection.CreateDocument(context.TODO(), food)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create food item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func UpdateFoodByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		foodID := chi.URLParam(r, "food_id")
		var food model.Food
		err := json.NewDecoder(r.Body).Decode(&food)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if food.Name != nil {
			updateObject["name"] = food.Name
		}
		if food.UnitPrice != nil {
			updateObject["unit_price"] = food.UnitPrice
		}
		if food.FoodImage != nil {
			updateObject["food_image"] = food.FoodImage
		}
		if food.MenuID != nil {
			var menu model.Menu
			_, err = menuCollection.ReadDocument(context.TODO(), *food.MenuID, &menu)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "menu was not"})
				return
			}
			updateObject["menu_id"] = food.MenuID
		}

		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = food.UpdatedAt

		meta, err := foodCollection.UpdateDocument(context.TODO(), foodID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create food item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteFoodByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		foodID := chi.URLParam(r, "food_id")

		meta, err := foodCollection.RemoveDocument(context.TODO(), foodID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete food item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}
