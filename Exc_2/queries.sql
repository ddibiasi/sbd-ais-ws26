-- 2.1

DROP TABLE IF EXISTS orders;

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_name TEXT,
    product_category TEXT, 
    quantity INT, 
    price_per_unit NUMERIC(10,2), 
    order_date DATE,
    country TEXT
);

\COPY orders(customer_name, product_category, quantity, price_per_unit, order_date, country) FROM '/data/orders_1M.csv' DELIMITER ',' CSV HEADER;

-- A
SELECT id
FROM orders
ORDER BY price_per_unit desc
LIMIT 1;

-- B
SELECT product_category, SUM(quantity)
FROM orders
GROUP BY product_category
ORDER BY SUM(quantity) desc
LIMIT 3;

-- C
SELECT product_category, SUM(price_per_unit * quantity) as revenue
FROM orders
GROUP BY product_category;

-- D
SELECT customer_name, sum(price_per_unit * quantity) as money_spent
FROM orders
GROUP BY customer_name
ORDER BY money_spent desc
LIMIT 5;

-- E 
SELECT customer_name, count(customer_name) as occurances
FROM orders
GROUP BY customer_name
ORDER BY occurances desc;


-- 2.2
EXPLAIN ANALYZE
SELECT COUNT(*)
FROM people_100k p1
JOIN people_100k p2
  ON p1.country = p2.country;


--- 2.2 3

SELECT SUM(country_count * country_count)
FROM (
    SELECT COUNT(country) AS country_count
    FROM people_100k
    GROUP BY country
);




