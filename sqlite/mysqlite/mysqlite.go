package mysqlite

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const CREATE_DB_FILE = "createTables.sql"

var (
	tableCreated = false
	Filename     = ""
)

type Order struct {
	ID         int64
	Customer   string
	Lines      []OrderLine
	OrderTotal float64
}

type OrderLine struct {
	ProductId            int
	Qty                  int
	ProductSellUnitPrice float64
	LineTotal            float64
}

func GetOrderTotal(lines []OrderLine) (orderTotal float64) {
	orderTotal = 0
	for i := 0; i < len(lines); i++ {
		orderTotal += lines[i].LineTotal
	}
	return
}

func FindOrder(orderId int64) (Order, error) {
	conn, err := openConnection()
	if err != nil {
		return Order{}, err
	}
	defer conn.Close()

	order := Order{}
	row := conn.QueryRow("select customer, total from my_order where order_id = ?", orderId)
	err = row.Scan(&order.Customer, &order.OrderTotal)
	if err != nil {
		if err == sql.ErrNoRows {
			return Order{}, errors.New("cant find the order")
		} else {
			return Order{}, err
		}
	}
	order.ID = orderId

	rows, err := conn.Query("select product_id, qty, product_sell_unit_price, total from my_order_lines where order_id = ?", orderId)
	if err != nil {
		return Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		line := OrderLine{}
		err = rows.Scan(&line.ProductId, &line.Qty, &line.ProductSellUnitPrice, &line.LineTotal)
		if err != nil {
			return Order{}, err
		}
		order.Lines = append(order.Lines, line)
	}

	return order, nil
}

// TODO. Just learning, this is NOT transactional
func AddOrder(order Order) (Order, error) {
	conn, err := openConnection()
	if err != nil {
		return order, err
	}
	defer conn.Close()

	res, err := conn.Exec("insert into my_order (customer, total) values (?, ?)",
		order.Customer,
		order.OrderTotal)
	if err != nil {
		return order, err
	}

	orderId, err := res.LastInsertId()
	if err != nil {
		return order, err
	}
	order.ID = orderId

	for _, line := range order.Lines {
		_, err := conn.Exec("insert into my_order_lines (order_id, product_id, qty, product_sell_unit_price, total) values (?, ?, ?, ?, ?)",
			order.ID,
			line.ProductId,
			line.Qty,
			line.ProductSellUnitPrice,
			line.LineTotal)
		if err != nil {
			return order, err
		}
	}

	return order, nil
}

func ExistOrder(orderId int64) (bool, error) {
	conn, err := openConnection()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	id := 0
	row := conn.QueryRow("select order_id from my_order where order_id = ?", orderId)
	err = row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func myInit(conn *sql.DB) error {
	if !tableCreated {
		sql, err := os.ReadFile(CREATE_DB_FILE)
		if err != nil {
			return err
		}
		_, err = conn.Exec(string(sql))
		if err != nil {
			return err
		}
		tableCreated = true
	}
	return nil
}

func openConnection() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", Filename)
	if err != nil {
		return nil, err
	}
	err = myInit(conn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
