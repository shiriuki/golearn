DROP TABLE IF EXISTS my_order;
DROP TABLE IF EXISTS my_order_lines;

CREATE TABLE my_order (
    order_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    customer VARCHAR(100) NOT NULL,
    total DECIMAL(7,2 )NOT NULL
);

/* TODO. Just learning. Add foreign keys */
CREATE TABLE my_order_lines (
    order_id Int NOT NULL,
    product_id Int NOT NULL,
    qty Int NOT NULL,
    product_sell_unit_price DECIMAL(5,2) NOT NULL,
    total DECIMAL(6,2) NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
