# Exercise 2: Operation Data Plane with OLTPs

Solutions for Activity 2.1 (E-commerce analytical queries) and Activity 2.2 (Why is the self-join so slow?). All queries were run on PostgreSQL 18.1 in the `pg-bigdata` Docker container.

---

## Activity 2.1: PostgreSQL Analytical Queries (E-commerce)

### Table definition

```sql
DROP TABLE IF EXISTS orders;

CREATE TABLE orders (
  customer_name    TEXT,
  product_category TEXT,
  quantity         INTEGER,
  price_per_unit   NUMERIC(10,2),
  order_date       DATE,
  country          TEXT
);

\COPY orders(customer_name,product_category,quantity,price_per_unit,order_date,country) FROM '/data/orders_1M.csv' DELIMITER ',' CSV HEADER
```

Type choices: `NUMERIC(10,2)` for prices so money is stored exactly (no floating point rounding), `INTEGER` for quantity, `DATE` for the order date, `TEXT` for names, categories and countries. The table holds 1,000,000 rows.

### A. Which order has the highest `price_per_unit`?

```sql
SELECT *
FROM orders
ORDER BY price_per_unit DESC
LIMIT 1;
```

```
 customer_name | product_category | quantity | price_per_unit | order_date | country
---------------+------------------+----------+----------------+------------+---------
 Emma Brown    | Automotive       |        3 |        2000.00 | 2024-10-11 | Italy
```



### B. Top 3 product categories by total quantity sold

```sql
SELECT product_category, SUM(quantity) AS total_quantity
FROM orders
GROUP BY product_category
ORDER BY total_quantity DESC
LIMIT 3;
```

```
 product_category | total_quantity
------------------+----------------
 Health & Beauty  |         300842
 Electronics      |         300804
 Toys             |         300598
```



### C. Total revenue per product category

```sql
SELECT product_category, SUM(price_per_unit * quantity) AS revenue
FROM orders
GROUP BY product_category
ORDER BY revenue DESC;
```

```
 product_category |   revenue
------------------+--------------
 Automotive       | 306589798.86
 Electronics      | 241525009.45
 Home & Garden    |  78023780.09
 Sports           |  61848990.83
 Health & Beauty  |  46599817.89
 Office Supplies  |  38276061.64
 Fashion          |  31566368.22
 Toys             |  23271039.02
 Grocery          |  15268355.66
 Books            |  12731976.04
```


### D. Top 5 customers by total spending

```sql
SELECT customer_name,
       SUM(price_per_unit * quantity) AS total_spent,
       COUNT(*) AS order_count
FROM orders
GROUP BY customer_name
ORDER BY total_spent DESC
LIMIT 5;
```

```
  customer_name  | total_spent | order_count
----------------+-------------+-------------
 Carol Taylor   |   991179.18 |        1028
 Nina Lopez     |   975444.95 |         980
 Daniel Jackson |   959344.48 |        1033
 Carol Lewis    |   947708.57 |         943
 Daniel Young   |   946030.14 |         973
```



### E. What do you notice, and why?

There are 32 first names and 33 last names. The random_customer_name() function generates 32 x 33 random names this equals to 1056 names generated. However , we have a million orders. 1 million / 1056 approx. equals to 947 orders PER NAME.The order_count values are all close to this average.

Check:

```sql
SELECT COUNT(DISTINCT customer_name) FROM orders;
```

This gives me 1056.

---

## Activity 2.2: Why is this self-join so slow?

The query:

```sql
SELECT COUNT(*)
FROM people_big p1
JOIN people_big p2
  ON p1.country = p2.country;
```

### Step 1: Measure how it grows

```sql
CREATE TABLE people_50k  AS SELECT * FROM people_big WHERE id <= 50000;
CREATE TABLE people_100k AS SELECT * FROM people_big WHERE id <= 100000;
CREATE TABLE people_200k AS SELECT * FROM people_big WHERE id <= 200000;
```

| rows in table | join result (`COUNT(*)`) | time |
|---|---|---|
| 50 000 | 27 501 822 | 5.9 s |
| 100 000 | 109 946 508 | 30.1 s |
| 200 000 | 439 395 606 | 137.6 s (2 min 17 s) |
| 1 000 000 (prediction) | about 10.98 billion | about 3 440 s (about 57 min) |
| 1 000 000 (actual result, from the Step 3 rewrite) | 10 983 941 260 | not run as a join |

**Growth factors when the input doubles**

| step | result grows by | time grows by |
|---|---|---|
| 50k to 100k | 109 946 508 / 27 501 822 = 4.0x | 30.1 / 5.9 = 5.1x |
| 100k to 200k | 439 395 606 / 109 946 508 = 4.0x | 137.6 / 30.1 = 4.6x |

When the rows double, the result grows by **4x**, and the time by roughly **4x to 5x**. This is quadratic growth. If a country has k people, the join pairs every person with every person of the same country (including themselves), so that country contributes k x k pairs. Doubling the table doubles every k, so the number of pairs grows 2 x 2 = 4 times. The time follows the number of pairs because the database has to produce and count every pair. The time grew slightly faster than 4x, probably because larger joins use more memory and disk work.

**Prediction for 1M rows.** 1M is 5 times larger than 200k, so the result grows by 5 x 5 = 25 times: 439 395 606 x 25 = about 10.98 billion. The time grows by about 25 times also : 137.6 s x 25 = about 3 440 s, which is about 57 minutes. This justifies the exercise statement ("more than 10 minutes"). The real result from Step 3 is 10 983 941 260, so the size prediction was very accurate.

### Step 2: Does an index help?

```sql
CREATE INDEX idx_people_100k_country ON people_100k(country);
ANALYZE people_100k;

SELECT COUNT(*) FROM people_100k p1 JOIN people_100k p2 ON p1.country = p2.country;

EXPLAIN ANALYZE
SELECT COUNT(*) FROM people_100k p1 JOIN people_100k p2 ON p1.country = p2.country;
```

Results:

| | result | time |
|---|---|---|
| without index | 109 946 508 | 30.1 s |
| with index on `country` | 109 946 508 | 16.1 s |

The index made the query about **2x faster**, but the result is still 109.9 million pairs and the query still has to produce and count every one of them.

**The index gives only a slight improvement.** An index is useful when a query needs a few rows out of many, because it avoids reading the whole table. Here there are only a small number of countries, so each person matches thousands of other people. There is nothing to skip. The index may make the lookup of matches a bit cheaper, which explains the constant-factor gain, but it cannot reduce the number of pairs. The cost is dominated by the output size, not by the process of finding rows.

### Step 3: Rewrite without a join

If a country has k people, it contributes k x k pairs. So the total number of pairs is the sum of k x k over all countries, and we only need the count of people per country:

```sql
SELECT SUM(people_per_country * people_per_country) AS pair_count
FROM (
  SELECT country, COUNT(*) AS people_per_country
  FROM people_100k
  GROUP BY country
) AS country_counts;
```

Result on `people_100k`: **109 946 508**, exactly the same as the join.

On the full table:

```sql
SELECT SUM(people_per_country * people_per_country) AS pair_count
FROM (
  SELECT country, COUNT(*) AS people_per_country
  FROM people_big
  GROUP BY country
) AS country_counts;
```

| | result | time |
|---|---|---|
| rewrite on `people_big` | 10 983 941 260 | **2.25 s** |
| predicted join on `people_big` | about 10.98 billion | about 3 440 s |

The rewrite is about **1 500 times faster** than the predicted join time (3 440 s / 2.25 s), and the result matches the predicted size. The rewrite reads the table once (linear work), while the join creates about 11 billion pairs (quadratic work).

### Step 4: Discussion

**1. What the rewrite tells us about hardware and indexes**


The self-join takes every person and matches them with every other person from the same country (including themselves). 
Then COUNT(*) counts how many matched pairs came out.To do that, the database actually creates every single pair, one by one. 
With 1M rows that is about 11 billion pairs. The re-write counts how many people per country [inner query] and then squares the count [outer query]. 
The size of the improvement came from changing the algorithm, not from using more resources. A better query is also much cheaper than more hardware, 
and its advantage grows as the data grows.


**2. What if the business needs the pairs themselves?**

Then the output has about 11 billion rows and no rewrite can avoid producing them. Storing only two 4-byte ids per pair would already take about 88 GB.

- *A bigger machine* helps only by a constant factor. Twice the cores gives at most twice the speed, and memory and disk have to handle the huge output as well.
- *A cluster* can split the work, for example by partitioning the data by country, so N nodes can finish in up to 1/N of the time. But there are limits: the work per country is k x k, so the largest country becomes the bottleneck while other nodes wait. The pairs must also be written and often sent over the network, which adds cost.
- *The growth stays the same.* With 5 times more data there are 25 times more pairs, so keeping the same runtime would need about 25 times more nodes. A cluster makes the job possible, but it does not make it scale well.

In practice I would first check whether all pairs are really needed. Options are sampling, filtering with an additional join condition (for example same country and same department), processing one country at a time, or streaming results to the consumer instead of storing them.

**3. Limits of an OLTP database for this workload, especially in the cloud**

PostgreSQL here is a single-node, row-oriented system designed for many small, fast transactions. For this workload:

- The query reads whole rows even though only `country` is needed, and it mostly runs in one process.
- It shares CPU, memory and disk with every other user, so one naive query can slow down the whole database.
- Scaling one machine up (more CPU and RAM) becomes expensive quickly and has a hard ceiling.

In a large-scale cloud environment the usual approach is to separate the workloads:

- Keep the OLTP database for transactions.
- Copy the data to an analytical system (a data warehouse such as BigQuery, Snowflake or Redshift) with columnar storage and distributed processing. There, compute and storage scale independently and heavy queries run on their own resources, so they do not slow down transactions.
- Smaller measures help too: read replicas for reporting, precomputed aggregates or materialized views, and safeguards such as `statement_timeout` to stop runaway queries.

Even with a distributed engine, a quadratic query stays quadratic. The cloud can reduce the cost of a badly designed query, but it cannot remove it. Query design stays the most important factor for scalability and efficiency.
