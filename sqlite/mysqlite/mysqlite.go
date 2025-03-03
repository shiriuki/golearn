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

func FindOrder(orderId int64) (Order, error) {
	conn, err := openConnection()
	if err != nil {
		return Order{}, err
	}
	defer conn.Close()

	order := Order{}
	_, err = oneRowQuery(
		conn, "select customer, total from my_order where order_id = ?", []any{orderId},
		[]any{&order.Customer, &order.OrderTotal}, true, "cant find the order")
	if err != nil {
		return Order{}, err
	}
	order.ID = orderId

	rowProcessor := func(currentRow *sql.Rows) (result any, err error) {
		line := OrderLine{}
		err = currentRow.Scan(
			&line.ProductId, &line.Qty, &line.ProductSellUnitPrice, &line.LineTotal)
		if err != nil {
			return nil, err
		}
		return line, nil
	}
	results, err := multiRowQuery(conn, "select product_id, qty, product_sell_unit_price, total from my_order_lines where order_id = ?", []any{orderId}, rowProcessor)
	if err != nil {
		return Order{}, err
	}
	order.Lines = convertToType[OrderLine](results)

	return order, nil
}

func ExistOrder(orderId int64) (orderExist bool, err error) {
	conn, err := openConnection()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return oneRowQuery(
		conn, "select order_id from my_order where order_id = ?", []any{orderId},
		[]any{&orderId}, false, "")
}

func convertToType[T any](input []any) []T {
	output := make([]T, 0, len(input))
	for i, _ := range input {
		obj, isOfType := input[i].(T)
		if isOfType {
			output = append(output, obj)
		}
	}
	return output
}

func multiRowQuery(
	conn *sql.DB, querySql string, queryParams []any,
	rowProcessor func(currentRow *sql.Rows) (result any, err error)) (results []any, err error) {
	rows, err := conn.Query(querySql, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results = []any{}
	for rows.Next() {
		result, err := rowProcessor(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func oneRowQuery(
	conn *sql.DB, querySql string, queryParams []any, results []any,
	noRowIsError bool, noRowErrMsg string) (rowFound bool, err error) {
	row := conn.QueryRow(querySql, queryParams...)
	err = row.Scan(results...)
	if err != nil {
		if err == sql.ErrNoRows {
			if noRowIsError {
				return false, errors.New(noRowErrMsg)
			} else {
				return false, nil
			}
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
