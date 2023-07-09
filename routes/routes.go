package routes

import (
	"main/controller"

	"github.com/go-chi/chi/v5"
)

func Use(router *chi.Mux) {
	router.Group(func(r chi.Router) {
		// food routes
		r.Route("/foods", func(r chi.Router) {
			r.Get("/", controller.GetFoods())
			r.Post("/", controller.CreateFood())
			r.Get("/{food_id}", controller.GetFoodByID())
			r.Patch("/{food_id}", controller.UpdateFoodByID())
			r.Delete("/{food_id}", controller.DeleteFoodByID())
		})

		// invoice routes
		r.Route("/invoices", func(r chi.Router) {
			r.Get("/", controller.GetInvoices())
			r.Post("/", controller.CreateInvoice())
			r.Get("/{invoice_id}", controller.GetInvoiceByID())
			r.Patch("/{invoice_id}", controller.UpdateInvoiceByID())
			r.Delete("/{invoice_id}", controller.DeleteInvoiceByID())
		})

		// menu routes
		r.Route("/menus", func(r chi.Router) {
			r.Get("/", controller.GetMenus())
			r.Post("/", controller.CreateMenu())
			r.Get("/{menu_id}", controller.GetMenuByID())
			r.Patch("/{menu_id}", controller.UpdateMenuByID())
			r.Delete("/{menu_id}", controller.DeleteMenuByID())
		})

		// order routes
		r.Route("/orders", func(r chi.Router) {
			r.Get("/", controller.GetOrders())
			r.Post("/", controller.CreateOrder())
			r.Get("/{order_id}", controller.GetOrderByID())
			r.Patch("/{order_id}", controller.UpdateOrderByID())
			r.Delete("/{order_id}", controller.DeleteOrderByID())
		})

		// table routes
		r.Route("/tables", func(r chi.Router) {
			r.Get("/", controller.GetTables())
			r.Post("/", controller.CreateTable())
			r.Get("/{table_id}", controller.GetTableByID())
			r.Patch("/{table_id}", controller.UpdateTableByID())
			r.Delete("/{table_id}", controller.DeleteTableByID())
		})

		// orderItem routes
		r.Route("/orderItems", func(r chi.Router) {
			r.Get("/", controller.GetOrderItems())
			r.Post("/", controller.CreateOrderItem())
			r.Get("/{orderItem_id}", controller.GetOrderItemByID())
			r.Get("/order/{order_id}", controller.GetOrderItemsByOrder())
			r.Patch("/{orderItem_id}", controller.UpdateOrderItemByID())
			r.Delete("/{orderItem_id}", controller.DeleteOrderItemByID())
		})
	})
}
