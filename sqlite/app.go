package main

import (
	"fmt"
	"log"
	"mecc/sqlite/mysqlite"
)

func main() {
	mysqlite.Filename = "sqlite3.db"
	exist, err := mysqlite.ExistOrder(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Order exist?", exist)

	lines := []mysqlite.OrderLine{
		{ProductId: 7, Qty: 2, ProductSellUnitPrice: 2.0, LineTotal: 4.0},
		{ProductId: 11, Qty: 1, ProductSellUnitPrice: 17.5, LineTotal: 17.5},
	}
	order := mysqlite.Order{
		Customer:   "Manuel Calvo",
		OrderTotal: mysqlite.GetOrderTotal(lines),
		Lines:      lines,
	}
	order, err = mysqlite.AddOrder(order)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Order created. Id", order.ID)

	exist, err = mysqlite.ExistOrder(order.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Order exist?", exist)

	order, err = mysqlite.FindOrder(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(order)
}
