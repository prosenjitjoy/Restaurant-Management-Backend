package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/database"
	"main/model"
	"math"
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var orderItemCollection = database.OpenCollection(db, "orderItems")

func GetOrderItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR orderItem IN orderItems LIMIT 10 RETURN orderItem"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		orderItems := []model.OrderItem{}
		for {
			var orderItem model.OrderItem
			_, err := cursor.ReadDocument(context.TODO(), &orderItem)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read order items"})
				return
			}

			orderItems = append(orderItems, orderItem)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderItems)
	}
}

func GetOrderItemByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderItemID := chi.URLParam(r, "orderItem_id")
		var orderItem model.OrderItem

		_, err := orderItemCollection.ReadDocument(context.TODO(), orderItemID, &orderItem)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch orderItem"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderItem)
	}
}

func CreateOrderItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orderItemPack model.OrderItemPack
		var order model.Order

		err := json.NewDecoder(r.Body).Decode(&orderItemPack)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		var table model.Table
		_, err = tableCollection.ReadDocument(context.TODO(), *orderItemPack.TableID, &table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch tableID"})
			return
		}

		order.OrderDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToBeInserted := []model.OrderItem{}
		order.TableID = orderItemPack.TableID
		orderID := OrderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = orderID
			err := validate.Struct(orderItem)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
				return
			}

			var food model.Food
			_, err = foodCollection.ReadDocument(context.TODO(), *orderItem.FoodID, &food)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to fetch food item"})
				return
			}

			orderItem.OrderItemID = uuid.NewString()
			orderItem.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			totalPrice := math.Round((*food.UnitPrice)*(*orderItem.Quantity)*100) / 100
			orderItem.TotalPrice = &totalPrice
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		metas, errs, err := orderItemCollection.CreateDocuments(context.TODO(), orderItemsToBeInserted)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create orderItem Collection"})
			return
		} else if err := errs.FirstNonNil(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create orderItem Collection"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(metas.Keys())
	}
}

func UpdateOrderItemByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderItemID := chi.URLParam(r, "orderItem_id")
		var orderItem model.OrderItem
		err := json.NewDecoder(r.Body).Decode(&orderItem)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if orderItem.TotalPrice != nil {
			updateObject["total_price"] = orderItem.TotalPrice
		}
		if orderItem.Quantity != nil {
			updateObject["quantity"] = orderItem.Quantity
		}
		if orderItem.FoodID != nil {
			var food model.Food
			_, err = foodCollection.ReadDocument(context.TODO(), *orderItem.FoodID, &food)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "food was not found"})
				return
			}
			updateObject["food_id"] = orderItem.FoodID
		}

		orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = orderItem.UpdatedAt

		meta, err := orderItemCollection.UpdateDocument(context.TODO(), orderItemID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create orderItem"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteOrderItemByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderItemID := chi.URLParam(r, "orderItem_id")

		meta, err := orderItemCollection.RemoveDocument(context.TODO(), orderItemID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete orderItem"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func GetOrderItemsByOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "order_id")

		allOrderItems, err := ItemsByOrder(orderID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "error occured while listing order items by order id"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(allOrderItems)
	}
}

func ItemsByOrder(orderID string) (orderItemsByOrder []model.OrderItemsByOrder, err error) {
	fmt.Println(orderID)
	query := fmt.Sprintf(`
	LET foodList = (
	FOR orderItem IN orderItems
		FILTER orderItem.order_id == '%s'
			FOR food IN foods
				FILTER food._key == orderItem.food_id
				RETURN {
					image: food.food_image,
					name: food.name,
					quantity: orderItem.quantity,
					unit_price: food.unit_price,
					total_price: orderItem.total_price
				}
	)
	FOR orderItem IN orderItems
		FILTER orderItem.order_id == '%s'
			FOR order IN orders
				FILTER order._key == orderItem.order_id
				FOR table IN tables
					FILTER table._key == order.table_id
					RETURN DISTINCT {
						total_count: length(foodList),
						table_number: table.table_number,
						order_items: foodList,
						payment_due: SUM(foodList[*].total_price)
					}	
	`, orderID, orderID)

	cursor, err := db.Query(context.TODO(), query, nil)
	if err != nil {
		log.Fatal("Failed to run aggregation:", err)
	}
	defer cursor.Close()

	for {
		var orderItemByOrder model.OrderItemsByOrder
		_, err = cursor.ReadDocument(context.TODO(), &orderItemByOrder)
		if driver.IsNoMoreDocuments(err) {
			return orderItemsByOrder, nil
		} else if err != nil {
			log.Fatal("failed to get all order items by order:", err)
		}

		orderItemsByOrder = append(orderItemsByOrder, orderItemByOrder)
	}
}
