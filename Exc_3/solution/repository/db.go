package repository

import "ordersystem/model"

type DatabaseHandler struct {
	drinks []model.Drink
	orders []model.Order
}

func NewDatabaseHandler() *DatabaseHandler {
	drinks := []model.Drink{
		{ID: 1, Name: "Espresso", Description: "Strong and small", Price: 2.5},
		{ID: 2, Name: "Latte", Description: "Milky coffee", Price: 3.8},
	}
	orders := []model.Order{}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {
	totalledOrders := make(map[uint64]uint64)
	for _, order := range db.orders {
		totalledOrders[order.DrinkID] += order.Amount
	}
	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	db.orders = append(db.orders, *order)
}
