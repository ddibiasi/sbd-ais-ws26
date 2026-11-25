# Activity 2.1 — PostgreSQL Analytical Queries (E-commerce)

```sql
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
```
### A
```sql
SELECT id
FROM orders
ORDER BY price_per_unit desc
LIMIT 1;
```

|id   |
--------
| 841292|

### B
```sql
SELECT product_category, SUM(quantity)
FROM orders
GROUP BY product_category
ORDER BY SUM(quantity) desc
LIMIT 3;
```
 
| product_category | sum    |
|------------------|--------|
| Health & Beauty  | 300842 |
| Electronics      | 300804 |
| Toys             | 300598 |

### C
```sql
SELECT product_category, SUM(price_per_unit * quantity) as revenue
FROM orders
GROUP BY product_category;
```
| product_category | revenue      |
|------------------|--------------|
| Automotive       | 306589798.86 |
| Books            | 12731976.04  |
| Electronics      | 241525009.45 |
| Fashion          | 31566368.22  |
| Grocery          | 15268355.66  |
| Health & Beauty  | 46599817.89  |
| Home & Garden    | 78023780.09  |
| Office Supplies  | 38276061.64  |
| Sports           | 61848990.83  |
| Toys             | 23271039.02  |

### D
```sql
SELECT customer_name, sum(price_per_unit * quantity) as money_spent
FROM orders
GROUP BY customer_name
ORDER BY money_spent desc
LIMIT 5;
```
| customer_name  | money_spent |
|----------------|-------------|
| Carol Taylor   | 991179.18   |
| Nina Lopez     | 975444.95   |
| Daniel Jackson | 959344.48   |
| Carol Lewis    | 947708.57   |
| Daniel Young   | 946030.14   |

### E
```sql
SELECT customer_name, count(customer_name) as occurances
FROM orders
GROUP BY customer_name
ORDER BY occurances desc
LIMIT 5;
```
| customer_name  | occurances |
|----------------|------------|
| Lucas Williams | 1043       |
| Ivy Young      | 1043       |
| Alice Sanchez  | 1040       |
| David Gonzalez | 1038       |
| Daniel Jackson | 1033       |

The spendings ber customer in part D is sp high because each customer name appears around 1000 times. As in the dataset_generator there are only 32 firstnames and 32 lastnames it is inevitable, that in 1 million generations there are no duplicates.


# Activity 2.2 — Why Is This Self-Join So Slow?

### 1

| rows in table | join result (`COUNT(*)`) | time |
|---|---|---|
| 50 000 | 27 501 822| 927.912 ms|
| 100 000 | 109 946 508| 4085.304 ms|
| 200 000 | 439 395 606| 20389.380 ms|
| 1 000 000 | 11 000 000 000| 6 min 40 sec|

The result size and the time grows quadratic to the rows in the table. So if the rows are doubled the result and the time gets 2^2 which is 4 times bigger.

1 000 000 is 10 times bigger than 100 000 which means a growth for time and result by 100. 4 seconds * 100 = 400 seconds = 6 minutes 40 seconds. And the size is approx. 11 000 000 000

### 2

```sql
CREATE INDEX idx_people_big_country ON people_100k (country);
```

With the index it went down form roughly 4 seconds to just 3 seconds.
The imporvement is not only little because we are still iterating through 110 000 000 pairs, but there is at least little imporvement as the presorted countries can be found faster.

In the explain analyze: It can be seen that the index is used (Parallel Index Only Scan using idx_people_big_country on people_100k p1)

### 3

```sql
SELECT SUM(country_count * country_count)
FROM (
    SELECT COUNT(country) AS country_count
    FROM people_big
    GROUP BY country
);
```

Both queries return the same number for people_100k: 

<strong>For the big table:</strong>
|| Join Result | Time |
|---|---|---|
|predicted| 11 000 000 000| 6 min 40 sec|
|new query| 10 983 941 260| 343.049 ms|

With the new query we managed to reduce the time from over 6 minutes to less than half a second.


### 4

<strong>What the rewrite in Step 3 tells you about adding more hardware or an index?</strong>

If the query is inefficient or the task is approached in a wrong way, adding an index or upgrading the hardware hardly brings an imporvement (at least not a significant one). Instead an huge improvement can be made by taking a different approach. So all in all, if an index or harware upgrade doesn't bring an significant improvement, maybe a different approach has to be considered.

<strong>what you would do if the business actually needed the pairs themselves (not just their count) — would a bigger machine or a 
cluster help, and how much?</strong>

If you really need all the pairs (and assuming the are all different), i think you dont really have a choice and you have to do the query with the join. A bigger machine and a cluster would definitely help but the imporvements would not be massive. The physical limits of dealing with 11 billion rows are the bottleneck here and this can't be imporoved. So only small imporvement possible.

<strong>the limits of an OLTP database for this workload, especially in a large-scale cloud environment.</strong>

As OLTP databases are built for single transactions (finding the needle in a haystack), so it will be as what we already did.
The memory will be exhausted and the CPU is again the bottleneck.

In the cloud environment, this would consume all the database resources, causing the system to lock up and crash the application for regular users. By upgrading the machine, we only scale linear, as we would ned quadratic this doesn't help much.

