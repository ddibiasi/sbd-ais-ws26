package rest

import (
	"encoding/json"
	"net/http"
	"ordersystem/model"
	"ordersystem/repository"

	"github.com/go-chi/render"
)

// GetMenu 			godoc
// @tags 			Menu
// @Description 	Returns the menu of all drinks
// @Produce  		json
// @Success 		200 {array} model.Drink
// @Router 			/api/menu [get]
func GetMenu(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		menu := db.GetDrinks()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, menu)
	}
}

// todo create GetOrders /api/order/all
func GetOrders(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders := db.GetOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, orders)
	}
}

// todo create GetOrdersTotal /api/order/total
func GetOrdersTotal(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the map of total drink frequencies from your database handler
		totals := db.GetTotalledOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, totals)
	}
}

// PostOrder 		godoc
// @tags 			Order
// @Description 	Adds an order to the db
// @Accept 			json
// @Param 			b body model.Order true "Order"
// @Produce  		json
// @Success 		200
// @Failure     	400
// @Router 			/api/order [post]
func PostOrder(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo
		// declare empty order struct
		// err := json.NewDecoder(r.Body).Decode(&<your-order-struct>)
		// handle error and render Status 400
		// add to db

		var newOrder model.Order

		err := json.NewDecoder(r.Body).Decode(&newOrder)

		if err != nil {
			render.Status(r, http.StatusBadRequest) // 400
			render.JSON(w, r, map[string]string{"error": "Invalid data"})
			return
		}

		db.AddOrder(&newOrder)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, "ok")
	}
}
