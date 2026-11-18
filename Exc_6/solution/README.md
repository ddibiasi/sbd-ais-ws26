# Exercise 6: Card Sentinel — Write-up

## Part 1: The Decision Paragraph
For this scenario, **Option 4 (Log-based CDC via Debezium)** is the correct choice. Because the payment path belongs to another certified team, we cannot modify `card_processor.py` (ruling out Dual Write and Outbox patterns). Option 1 (Polling) imposes heavy query load on the system of record and risks missing out-of-order commits. Log-based CDC reads PostgreSQL's write-ahead log (WAL) entirely outside the application layer, satisfying **R1** (system of record untouched) and **R2** (every payment delivered reliably in order) with zero application code changes. Its primary cost is a replication slot that retains WAL data if consumers lag or crash, risking disk space exhaustion if unmonitored. We would monitor disk usage and replication lag via Prometheus/Grafana to catch issues before disk saturation stops the database. If our team owned the payment path, we would choose **Option 3 (Outbox pattern)** for native transactional safety.

---

## Part 2: The Pipeline in Three Runs

### Run 1 — Debezium's Defaults
* **Prediction:** All three velocity instances will not share work, and **R4 will fail**. Debezium defaults to creating a topic with 1 partition, which can only be consumed by a single instance within a consumer group, leaving the other instances idle and lagging behind the 1,000 payments/sec rate.
* **Scoreboard Results & Observation:** Only `vel-1` received assigned partitions (`[0]`), while `vel-2` and `vel-3` remained idle. The single instance became overloaded, lag grew continuously, and **R4 failed**.
* **Explanation:** Without explicit configuration, Debezium created a 1-partition topic. Kafka consumer groups assign each partition to exactly one consumer, capping parallelism at 1.

### Run 2 — More Partitions
* **Prediction:** All three instances will work and handle the load (**R4 passes**), but the velocity model will **fail R3** (catching almost no frauds). Because the default event key is the table's primary key (`id`), a cardholder's rapid test payments will be hashed and scattered randomly across all 6 partitions.
* **Scoreboard Results & Observation:** **R4 passed** as all three velocity instances received partitions and kept up with the throughput. However, **R3 failed** because the velocity model caught almost 0 of the card-testing frauds.
* **Explanation:** Keying by primary key (`id`) spreads a single cardholder's test transactions across different partitions and agent instances. No single instance saw enough sequential payments within its 60-second window to cross the threshold of 8.

### Run 3 — The Cardholder as the Key
* **Prediction:** The event key in Kafka will change to `{"user_id": ...}`. All payments belonging to the same cardholder will hash to the exact same partition, routed to the same velocity instance in strict commit order. Both **R3 and R4 will pass**.
* **Scoreboard Results & Observation:** **R2, R3, and R4 all passed**. The velocity model successfully detected card-testing attacks while all instances maintained real-time processing throughput.
* **Explanation:** Setting `message.key.columns` to `user_id` ensures proper stateful aggregation, routing all transactions for a specific user sequentially to the same agent instance.

---

## Closing Answer
Debezium chose defaults (single partitioning and primary-key routing) that are optimal for simple database table replication, but inadequate for stateful real-time analytics. If left unchecked in production, the system would not crash or throw explicit errors; instead, it would silently fail its business objectives by bottlenecking processing lag (R4) and blinding the velocity model to fraud rings (R3). The team could have noticed this by monitoring business-level metrics (sudden drop in fraud detection rates), tracking consumer lag in Kafka UI, and validating performance against a ground-truth scoreboard.
