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

var invoiceCollection = database.OpenCollection(db, "invoices")

func GetInvoices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "FOR invoice IN invoices LIMIT 10 RETURN invoice"
		cursor, err := db.Query(context.TODO(), query, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to execute query"})
			return
		}
		defer cursor.Close()

		invoices := []model.Invoice{}
		for {
			var invoice model.Invoice
			_, err := cursor.ReadDocument(context.TODO(), &invoice)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(status{"error": "failed to read menu items"})
				return
			}

			invoices = append(invoices, invoice)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invoices)
	}
}

func GetInvoiceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceID := chi.URLParam(r, "invoice_id")
		var invoice model.Invoice

		_, err := invoiceCollection.ReadDocument(context.TODO(), invoiceID, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch invoice item"})
			return
		}

		var invoiceView model.InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.OrderID)
		if err != nil || len(allOrderItems) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "error occured while listing order items by order id"})
			return
		}
		invoiceView.OrderID = invoice.OrderID
		invoiceView.PaymentDueDate = invoice.PaymentDueDate

		invoiceView.PaymentMethod = "null"
		if invoice.PaymentMethod != nil {
			invoiceView.PaymentMethod = *invoice.PaymentMethod
		}

		invoiceView.InvoiceID = invoice.InvoiceID
		invoiceView.PaymentStatus = invoice.PaymentStatus
		invoiceView.PaymentDue = allOrderItems[0].PaymentDue
		invoiceView.TableNumber = allOrderItems[0].TableNumber
		invoiceView.OrderDetails = allOrderItems[0].OrderItems

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(invoiceView)
	}
}

func CreateInvoice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var invoice model.Invoice
		err := json.NewDecoder(r.Body).Decode(&invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		var order model.Order
		_, err = orderCollection.ReadDocument(context.TODO(), invoice.OrderID, &order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to fetch order item"})
			return
		}

		paymentStatus := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &paymentStatus
		}

		invoice.PaymentDueDate, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))

		invoice.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.InvoiceID = uuid.NewString()

		err = validate.Struct(invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "failed to validate json:"})
			return
		}

		meta, err := invoiceCollection.CreateDocument(context.TODO(), invoice)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to create invoice item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func UpdateInvoiceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceID := chi.URLParam(r, "invoice_id")
		var invoice model.Invoice
		err := json.NewDecoder(r.Body).Decode(&invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status{"error": "invalid json format:"})
			return
		}

		updateObject := make(map[string]interface{})

		if invoice.PaymentMethod != nil {
			updateObject["payment_method"] = invoice.PaymentMethod
		}
		if invoice.PaymentStatus != nil {
			updateObject["payment_status"] = invoice.PaymentStatus
		}

		paymentStatus := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &paymentStatus
		}

		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObject["updated_at"] = invoice.UpdatedAt

		meta, err := invoiceCollection.UpdateDocument(context.TODO(), invoiceID, updateObject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to update invoice item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}

func DeleteInvoiceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceID := chi.URLParam(r, "invoice_id")

		meta, err := invoiceCollection.RemoveDocument(context.TODO(), invoiceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(status{"error": "failed to delete invoice item"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta.Key)
	}
}
