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

var orderCollection = database.OpenCollection(db, "orders")

func GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR order IN orders LIMIT 10 RETURN order"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		orders := []model.Order{}
		for {
			var order model.Order
			_, err := cursor.ReadDocument(context.TODO(), &order)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read order items"})
				return
			}

			orders = append(orders, order)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orders)
	}
}

func GetOrderByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "order_id")
		var order model.Order

		_, err := orderCollection.ReadDocument(context.TODO(), orderID, &order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch order item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	}
}

func CreateOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var order model.Order
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		err = validate.Struct(order)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
			return
		}

		var table model.Table
		_, err = tableCollection.ReadDocument(context.TODO(), *order.TableID, &table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "table was not found"})
			return
		}

		order.OrderDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.OrderID = uuid.NewString()

		meta, err := orderCollection.CreateDocument(context.TODO(), order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create order item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func UpdateOrderByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "order_id")
		var order model.Order
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if order.TableID != nil {
			var table model.Table
			_, err = tableCollection.ReadDocument(context.TODO(), *order.TableID, &table)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "table was not found"})
				return
			}
			updateObject["table_id"] = order.TableID
		}

		order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = order.UpdatedAt

		meta, err := orderCollection.UpdateDocument(context.TODO(), orderID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create order item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteOrderByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "order_id")

		meta, err := orderCollection.RemoveDocument(context.TODO(), orderID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete order item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func OrderItemOrderCreator(order model.Order) string {
	order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.OrderID = uuid.NewString()
	_, err := orderCollection.CreateDocument(context.TODO(), order)
	if err != nil {
		panic(err)
	}
	return order.OrderID
}
