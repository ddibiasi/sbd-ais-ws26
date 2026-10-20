package repository

import (
	"ordersystem/model"
	"time"
)

type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// todo
func NewDatabaseHandler() *DatabaseHandler {
	// Init the drinks slice with some test data
	drinks := []model.Drink{
		{
			ID:          1,
			Description: "0.5 l from a happy cow",
			Name:        "Milk",
			Price:       3.5,
		},
		{
			ID:          2,
			Description: "0.5 l",
			Name:        "Beer",
			Price:       5,
		},
		{
			ID:          3,
			Description: "1/8 l",
			Name:        "Wine",
			Price:       5.25,
		},
		{
			ID:          4,
			Description: "0.5 l",
			Name:        "Apple Juice",
			Price:       1.2,
		},
		{
			ID:          5,
			Description: "0.3 l can",
			Name:        "Red Bull",
			Price:       1.5,
		},
	}
	// Init orders slice with some test data

	orders := []model.Order{
		{
			DrinkID:   2,
			Amount:    4,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			DrinkID:   1,
			Amount:    1,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			DrinkID:   4,
			Amount:    10,
			CreatedAt: time.Now().Add(-20 * time.Minute),
		},
		{
			DrinkID:   3,
			Amount:    2,
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			DrinkID:   2,
			Amount:    1,
			CreatedAt: time.Now().Add(-90 * time.Minute),
		},
		{
			DrinkID:   5,
			Amount:    3,
			CreatedAt: time.Now().Add(-15 * time.Minute),
		},
	}

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

// todo
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {

	totalledOrders := map[uint64]uint64{}

	for _, item := range db.orders {
		totalledOrders[item.DrinkID] += item.Amount
	}

	// calculate total orders
	// key = DrinkID, value = Amount of orders
	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// todo
	// add order to db.orders slice
	db.orders = append(db.orders, *order)
}
