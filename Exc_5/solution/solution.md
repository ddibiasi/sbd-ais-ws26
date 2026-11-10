# Exercise 5
Marco Hartl

## Part 2, Stage A

### Compare readings of the detector with the table
<strong>Detector: </strong> read 21,769 readings, last id 25,939, 4,155 readings missing, 517 anomalies

<strong>Table:  </strong> count: 25924
-> The Detector misses some data that is in the table

### Are the missing readings in the table?
Yes, the missing readings are in the table.

```python
with conn.transaction(), conn.cursor() as cur:
                        cur.executemany(
                            "INSERT INTO readings (run, machine_id, seq, ts, vibration, temperature, current) "
                            "VALUES (%(run)s, %(machine_id)s, %(seq)s, %(ts)s, %(vibration)s, %(temperature)s, %(current)s)",
                            batch)
                        time.sleep(rng.uniform(0, 0.12)) 
```

With cur.executemany() the data is generated and instantly written to a transaction in postgres.
The transaction is committed after the with block is exited, after the sleep function finishes.

### Why did the detector never see them?
Gateway 1: Ids: 100-114, delay: 100ms
Gateway 2: Ids: 115-129, delay: 0ms

The detector can only see committed data.
As the transaction with the data is committed with a delay (because of the sleep function) and the detector moves forward, it misses too late committed data.
Because there are other commits in the meantime (Gateway 2 with 115-129), the detector will then only look for id > 129.

### Which rows does the detector see when it polls in between, and which id does it remember?
If the detector polls between the commit of Gateway 1 and Gateway 2, It only sees the data from Gateway 2, sets its highest id to 129 and therfore completly misses Gateway 1.
This can happen regularly as the sleep function is random between 0 and 120 milliseconds.

### Watch the detector's lines while the plant runs: What happens to delay, backlog and missing, and why?
<strong>Delay: </strong> Fluctuates between 0.0 and 0.3 seconds. As it measures the time it takes between being inserted and fetched by the detector, the times fluctuate because there is a random sleep component in the insertion process and the detector sleeps for 0.2 seconds if there are no rows found.

<strong>Backlog: </strong> Spikes to 15-40 rows occasionally. The detector is fast enough (2000 rows per poll) to keep up with 20 machines. The backlog-metric shows some delay, because the metric sometimes gets printed at a time where the detector is about to process the data. (Difference between the highest processed id by the detector and the highest existing id)

<strong>Missing: </strong> Grows, as the detector is missing some data. (explained above)


## Part 2, Stage B

| | Requirement |
|---|---|
| R1 | **Scale:** 600 readings/s in the lab; the design must grow to 100,000 readings/s. |
| R2 | **Complete and ordered:** the model gets *every* reading of a machine, in `seq` order (the model counts `missing` and `out-of-order` readings). |
| R3 | **Fresh:** anomalies at most 5 s (p95) after the reading, at full rate. |
| R4 | **Independent consumers:** the raw readings also feed a training-data archive and a live dashboard. Adding a consumer must not slow down, change or break the others, and must not add load to PostgreSQL. |

<strong>Options</strong>
1. keep PostgreSQL, but poll it better (e.g. several pollers, each for a
   part of the machines);
2. the gateways write every reading into Kafka (or another log-based
   platform) instead of into PostgreSQL.

### Requirement Comparison

<ul>
    <li><strong>R1 Scale</strong></li>
    <ul>
        <li><strong>PostgreSQL (Risk): </strong>One PostgreSQL table that has to handle 100,000 inserts per second and continuous SELECT queries, will result in CPU and lock bottlenecks. Would only work good in the lab with 600 rows/s.</li>
        <li><strong>Kafka (Meets): </strong>Kafka can easily deal with huge and highly frequent inserts, so 100,000 rows/s are not a problem at all.</li>
    </ul>
    <li><strong>R2 Complete and Ordered</strong></li>
    <ul>
        <li><strong>PostgreSQL (Risk): </strong>As seen in the previous stage, the completeness didn't even work with smaller sizes. We could possibly wait for the older data to be committed and therefore meet the completeness requirement but then the time-distance to the new data gets bigger and bigger.</li>
        <li><strong>Kafka (Meets): </strong>If the machine_id is used as the key for Kafka, it guarantees that every reading for a machine lands in the same partition in chronological order.</li>
    </ul>
    <li><strong>R3 Fresh</strong></li>
    <ul>
        <li><strong>PostgreSQL (Risk): </strong>The data needs to be polled and this already introduces latency. Poll every 2 seconds and there are only 3 seconds left.</li>
        <li><strong>Kafka (Meets): </strong>The subscribed consumers can read the data in real-time (milliseconds)</li>
    </ul>
    <li><strong>R4 Independent Consumers</strong></li>
    <ul>
        <li><strong>PostgreSQL (Risk): </strong>Every consumer needs to do its own polling. So with every new consumer there is another SELECT query needed.</li>
        <li><strong>Kafka (Meets): </strong>With Consumer Groups the data is written once and then can be read by many. So no more load with more consumers.</li>
    </ul>    
</ul>


### Decision
In every requirement Kafka is just more fitting and performs (significantly) better in all of them, so we keep PostgreSQL just as a system of record. PostgreSQL fails especially for R4 as it simply cannot meet this requirement. If there are only a small amount of new rows per second, PostgreSQL would be better as it is easier.

### Amount of Detectors
1 reading every 3ms -> 333.34 readings/s
<ul>
    <li><strong>600 readings/s: </strong> 600/333.34 = 1.8 ~ 2 -> 2 detectors needed</li>
    <li><strong>100,000 readings/s</strong> 100,000/333.34 = 300 detectors needed.</li>
</ul>

### Tipping Point
<strong>Assumptions:</strong>
<ul>
    <li>standard SSD: 10,000 IOPS</li>
    <li>1 reading = 5 IOPS of writing to the database</li>
    <li>Poll for PostgreSQl every 2 seconds</li>
    <li>Poll 4 times, for archive, dashboard, 2 detectors</li>
</ul>

With 1000 reading/s: 1000 * 5 =  5000 IOPS

5000 IOPS for just writing to the database combined with the polling for 4 comsumers will exceed the limit of 10000.

So the tipping point is around 1000 readings/s.
