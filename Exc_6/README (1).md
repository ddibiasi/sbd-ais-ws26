# Exercise 6: Card Sentinel (Manuth)

## Part 1: Decision

I choose log-based CDC with Debezium. The payment path is certified and belongs to another team, so I may not change it (R1). CDC reads the write-ahead log, so it needs no code change and adds no queries on the system of record. It captures every committed payment in commit order, so each model sees every payment (R2), and adding another consumer does not change what the others get. The cost is the replication slot: PostgreSQL keeps all WAL that the connector has not read yet, so a connector that is stopped for a weekend can fill the disk and stop the database and the payment path with it. I would notice this early by monitoring the connector status, the retained WAL size of the slot and the consumer lag, with an alert well before the disk is full. If my team owned the payment path, I would use the outbox pattern, because it is also correct and complete, and I would avoid dual write because a crash between the commit and the send loses events.

## Run 1: Debezium defaults

Prediction: All three velocity instances share the work and R4 passes, because Debezium creates the topic and the group should split the partitions between three instances.

Scoreboard (velocity):

```
R3  frauds caught          0/2  (0%)   FAIL  (>= 90 %)
R3  false alerts           0   PASS
R4  no alerts and no telemetry   FAIL
```

Agents output: `vel-1 assigned partitions [0]`, `vel-2 assigned partitions []`, `vel-3 assigned partitions []`.

Note: in this run I started the agents after I had already stopped the card processor, so the agents worked on a backlog and the numbers above are not a clean measurement. The partition assignment is still valid.

Why: My prediction was wrong. Nobody told Debezium how many partitions to use, so the topic was created with 1 partition. Inside a consumer group one partition is read by only one instance, so vel-1 got partition 0 and vel-2 and vel-3 got nothing. The one busy instance cannot handle 1,000 payments per second, so its lag grows and R4 fails.

## Run 2: 6 partitions

Prediction: All three instances now work and R4 passes, but R3 fails, because the key of an event is still the primary key id.

Scoreboard (velocity): not re-measured. Expected: R4 PASS, R3 FAIL (0 or 1 of 4 frauds caught).

Why: With 6 partitions each instance gets two partitions (`[0, 1]`, `[2, 3]`, `[4, 5]`) and keeps up. The partition is hash(key) mod 6 and the key is the primary key, so the 12 test payments of one stolen card are spread over different partitions and instances. Each instance sees only about 4 of them within 60 seconds, below the threshold of 8, so no instance sees the whole attack.

## Run 3: cardholder as key

Prediction: The key becomes `{"user_id": ...}`, so R3 and R4 both pass.

Scoreboard (velocity): not re-measured. Expected: R3 PASS (at least 90 % of frauds caught), R4 PASS.

Why: With `message.key.columns` set to `public.transactions:user_id`, all payments of one cardholder go to the same partition, are read by the same instance and arrive in commit order. The velocity model sees the whole attack, so R3 passes, and the three instances still share the load, so R4 passes. The key has to be the entity whose state the model keeps, which is the cardholder and not the primary key of the table.

## Closing answer

Debezium chose two defaults that I had to change: one partition and the primary key as event key. If nobody had checked them, nothing would have crashed and no error would have appeared. First one instance would have fallen behind while the others sat idle, and after adding partitions the frauds of one card would have been split across instances and missed, so the fraud model would simply have caught fewer frauds. The team could notice this with consumer lag monitoring, idle instances in the consumer group, and a regular check of the catch rate against known frauds.
