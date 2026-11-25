# Exercise 6
Marco Hartl

## Part 1

### Which option do you choose for this scenario, and why?
- **1 Polling:** Fails, because it can miss payments (R2) and it also adds extra query load (R1)
- **2 Dual Write:** Fails, because the application gets changed (R1) and events can be missed or even new events can be hallucinated.
- **3 Outbox:** Fails, because it changes the application.
- **4 Log Based CDC:** Works best for this scenario as it guarantees that there is no data lost and it also doesn't change the application.

Considering that only CDC fits all the requirements, I would definetly choose CDC if I am not allowed to change the process.

### What does it cost, and how would you notice in production before it hurts?
**Log Based CDC:** The cost is, that PostgreSQL keeps all the data (files) in the WAL as long as the connector has not read it. It is only allowed to throw them away, if the connector has confirmed it has read it. So if the connector stops for a longer time, the disk of the WAL can fill up and this would cause the database and as a result also the payment process to stop.

To notice it before it hurts, a warning system would be needed. For example, setting up an alert if the disk space of the WAL exceeds a certain threshold (like 80%).

### Which option would you choose if your team owned the payment path?
If my team would own the process, then i would choose the Outbox method. Because with this method we don't have the risk of the WAL disk filling up and therefore stopping the whole payment. By using outbox we have one extra row in the transaction that pushes the data to the outbox table where the relay reads it and then forwards the event to kafka.


## Part 2

### Run 1

**Predictions: Do all three velocity instances share the work, and does R4 pass?**
By default Debrezium creates a kafka topic with only one partition. Kafka can only assign the partition to one consumer within the consumer group. So the prediction is, that only one of the velocity instances actually works and the other two just idle and do nothing. As we only have one velocity instance working, R4 will not pass as the workload is too huge for one instance.

**Kafka UI**
There is only 1 partition. By default, Debezium uses the primary key columns of the source database table as the Kafka message key.

**Agent**
```
[ano-1] assigned partitions [0]
[vel-2] assigned partitions []
[vel-3] assigned partitions []
[vel-1] assigned partitions [0]
```

Only one velocity instance gets the one partition.

The lag for the velocity models increases over time and gets more and more.


**Scoreboard**
```
velocity model (card testing)
R3  frauds caught          8/19  (42%)   FAIL  (>= 90 %)
R3  false alerts           1   PASS  (at most 10 % of the flagged transactions)
R4  alert latency p50/p95  105.3 s / 166.2 s   FAIL  (p95 <= 5 s after the commit)
    duplicate alerts       0   (info: alerts case management got twice)

anomaly model (account takeover)
R3  frauds caught          19/19  (100%)   PASS  (>= 90 %)
R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
R4  alert latency p50/p95  0.7 s / 0.9 s   PASS  (p95 <= 5 s after the commit)
    duplicate alerts       0   (info: alerts case management got twice)
```


**What happened and why?**
As we didn't tell Debrezium how many partitions to create, it defaultet to one partition. As Kafka can only assign the partition to one consumer instance, only one of the tree instances gets the partition. Therefore the other two idle and do nothing. But to deal with the load we need three running instances, as only one is running, it can't keep up with the work and therfore R4 fails.

### Run 2

**Predictions: Now all three instances work. Will the velocity model catch ≥ 90 % of the card-testing frauds (R3)?**
Now we have 6 partitions, so every velocity instance gets 2 partitions, this also means that the data is split between them. The data is split randomly across this 6 partitions meaning that the velocity instances don't get the connecting data samples. The connecting data samples are split and therefore the models will definitely not catch the 90% as the data is split across the instances.

**Kafka UI**
There are 6 partitions and the key is the same as before, the default of the database (payment id).

**Agent**
```
[ano-1] assigned partitions [0, 1, 2, 3, 4, 5]
[vel-1] assigned partitions [0, 1]
[vel-2] assigned partitions [2, 3]
[vel-3] assigned partitions [4, 5]
```

The partitions are evenly split between the instances and every instance now has work and runs.

The lag is constantly under 1 with max of 1.2 seconds.


**Scoreboard**
```
velocity model (card testing)
R3  frauds caught          0/16  (0%)   FAIL  (>= 90 %)
R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
R4  no alerts; worst lag of the velocity agents 1.2 s   PASS  (<= 5 s)
    duplicate alerts       0   (info: alerts case management got twice)

anomaly model (account takeover)
R3  frauds caught          16/16  (100%)   PASS  (>= 90 %)
R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
R4  alert latency p50/p95  0.7 s / 1.0 s   PASS  (p95 <= 5 s after the commit)
    duplicate alerts       0   (info: alerts case management got twice)
```


**What happened and why?**
The data from one cardholder is split up between the velocity instances because the databse standard key (payment id) is used. So the instance only gets about a third of the actual data and therefore doesn't recognize the fraud.


### Run 3

**Predictions: What will the key of an event look like in the Kafka UI, and what changes on the scoreboard?**
The key in the UI will show the key that we defined: the user_id. On the scoreboard we should now see that everything passes and everything works as intended as we now fix the last remaining problem.

**Kafka UI**
There are 6 partitions and the key is now the user_id.

**Agent**
```
[ano-1] assigned partitions [0, 1, 2, 3, 4, 5]
[vel-3] assigned partitions [4, 5]
[vel-1] assigned partitions [0, 1]
[vel-2] assigned partitions [2, 3]
```

The partitions are evenly split between the instances and every instance now has work and runs.

The lag is constantly under 1 with max of 1.2 seconds.


**Scoreboard**
```
velocity model (card testing)
R3  frauds caught          10/10  (100%)   PASS  (>= 90 %)
R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
R4  alert latency p50/p95  0.7 s / 1.0 s   PASS  (p95 <= 5 s after the commit)
    duplicate alerts       0   (info: alerts case management got twice)

anomaly model (account takeover)
R3  frauds caught          10/10  (100%)   PASS  (>= 90 %)
R3  false alerts           0   PASS  (at most 10 % of the flagged transactions)
R4  alert latency p50/p95  0.6 s / 0.9 s   PASS  (p95 <= 5 s after the commit)
    duplicate alerts       0   (info: alerts case management got twice)
```



**What happened and why?**
All the data from one user is now in one partition and with that it also gets to one velocity instance. The instance now has all the data it needs to say if this is fraud or not. 



### Debezium chose two defaults that you had to change. What would have happened in production if nobody had checked them, and how could the team have noticed?
If nobody checked them, then the setup would run like in run 1. Only one veloctiy instance would run and it could not keep up with detecting the frauds, so the delay would increase and in the end a fraud will be detected minutes after it happened.

Additionaly if someone would add more partitions, then the problem with the wrong key and therfore the mixed distribution of the data comes into play.

The team can notice by looking at the results and the delay, or in general at anything that shows the result or/and the delay. Then the team would immediately see that something is not right.



# Optional Part


## Step 1 — A replicated topic

```
Topic: payments TopicId: Ve0F8624SOqwEPWk3Qhpag PartitionCount: 3       ReplicationFactor: 3    Configs: min.insync.replicas=2
        Topic: payments Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3,1,2      Elr:    LastKnownElr: 
        Topic: payments Partition: 1    Leader: 1       Replicas: 1,2,3 Isr: 1,2,3      Elr:    LastKnownElr: 
        Topic: payments Partition: 2    Leader: 2       Replicas: 2,3,1 Isr: 2,3,1      Elr:    LastKnownElr:
```


## Step 2 — Write with acks=all
```
payments:0:0
payments:1:0
payments:2:4
```

All messages in one partition

## Step 3 — One broker fails
```
Topic: payments Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3,2        Elr:    LastKnownElr: 
Topic: payments Partition: 1    Leader: 2       Replicas: 1,2,3 Isr: 2,3        Elr:    LastKnownElr: 
Topic: payments Partition: 2    Leader: 2       Replicas: 2,3,1 Isr: 2,3        Elr:    LastKnownElr: 
```
### Produce again with acks=all (step 2) and count the records.

```
payments:0:11
payments:1:0
payments:2:4
```

All the new messages go to partition 0


### Who leads partition 0 now? How long did the failover take? Did the producer notice anything?

Broker 3 leads partition 0 now.

Takeover was immediately

The producer didn't notice as: min.insync.replicas=2


## Step 4 — Two brokers fail

```
Topic: payments Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3  Elr: 2  LastKnownElr: 
Topic: payments Partition: 1    Leader: 3       Replicas: 1,2,3 Isr: 3  Elr: 2  LastKnownElr: 
Topic: payments Partition: 2    Leader: 3       Replicas: 2,3,1 Isr: 3  Elr: 2  LastKnownElr: 
```

Broker 3 now leads every partition.


### produce with acks=1 (--producer-property acks=1); does the producer report an error?
No as it only needs a conformation from the leading broker.


### count the records (step 2) and read the topic (Ctrl+C to stop):

payments:0:11
payments:1:0
payments:2:4

The new messages are not there

### Where are your acks=1 records? Look inside broker-3's log files:
They are stored locally in broker-3. Because it was accepted after it wrote the data to itself.

### The difference between written to the leader's log and visible to consumers. Compare the offsets of step 4.3 with what you find in the log files.
Written to the leader's log: The leader broker records the acks=1 messages directly to its local disk file.  

Visible to consumers: Consumers cannot see them because the High Watermark remains frozen until followers come back online and replicate the data. 

## Step 5 — Recovery

### broker-2
```
Topic: payments Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3,2        Elr:    LastKnownElr: 
Topic: payments Partition: 1    Leader: 3       Replicas: 1,2,3 Isr: 3,2        Elr:    LastKnownElr: 
Topic: payments Partition: 2    Leader: 3       Replicas: 2,3,1 Isr: 3,2        Elr:    LastKnownElr: 
```
```
payments:0:11
payments:1:5
payments:2:4
```
All messages are there.

#### What happened to the acks=1 records of step 4?
When broker-2 started back up, it caught up by replicating those acks=1 records from the leader (broker-3), meaning they were safely copied to the newly recovered in-sync replica.   

#### Which component decided that they may now be read?
The leader broker (broker-3) decided they could now be read by advancing the High Watermark (HW) once the followers rejoined the In-Sync Replicas (ISR) and confirmed replication.

### broker-1

After starting broker-1 and rebalincing the leaders are back as before.

```
Topic: payments Partition: 0    Leader: 3       Replicas: 3,1,2 Isr: 3,2,1      Elr:    LastKnownElr: 
Topic: payments Partition: 1    Leader: 1       Replicas: 1,2,3 Isr: 3,2,1      Elr:    LastKnownElr: 
Topic: payments Partition: 2    Leader: 2       Replicas: 2,3,1 Isr: 3,2,1      Elr:    LastKnownElr: 
```

## Step 6 — The price of durability

### acks = 0

300000 records sent, 275229.4 records/sec (52.50 MB/sec), 172.79 ms avg latency, 271.00 ms max latency, 181 ms 50th, 230 ms 95th, 235 ms 99th, 239 ms 99.9th.

### acks = 1
300000 records sent, 249376.6 records/sec (47.56 MB/sec), 244.43 ms avg latency, 339.00 ms max latency, 259 ms 50th, 324 ms 95th, 334 ms 99th, 338 ms 99.9th.

### acks = all
300000 records sent, 173913.0 records/sec (33.17 MB/sec), 457.79 ms avg latency, 759.00 ms max latency, 447 ms 50th, 733 ms 95th, 752 ms 99th, 757 ms 99.9th.


With acks = all we loose performance but get saftey.


## Results sheet


| Step | Brokers running | Leaders (P0, P1, P2) | ISR of P0 | `acks=all` accepted? | `acks=1` accepted? | `acks=1` records readable? |
|---|---|---|---|---|---|---|
| 2 | 1, 2, 3 | 3, 1, 2 | 3, 1, 2 | Yes | – | – |
| 3 | 2, 3 | 3, 2, 2 | 3, 2 | Yes | – | – |
| 4 | 3 | 3, 3, 3 | 3 | No | Yes | No |
| 5 | 2, 3 | 3, 3, 3 | 3, 2 | – | – | Yes |

| `acks` | records/sec | avg latency (ms) | 99th percentile (ms) |
|---|---|---|---|
| 0 | 275,229.4 | 172.79 | 235 |
| 1 | 249,376.6 | 244.43 | 334 |
| all | 173,913.0 | 457.79 | 752 |


## Check yourself

### With RF 3 and `min.insync.replicas=2`, how many brokers can fail before `acks=all` producers get errors? And before data is lost?
One broker can fail before acks=all producers get errors, and two can fail before data is lost.

### Why did `acks=1` look successful in step 4, and why is that misleading?
It looked successful because the single surviving leader accepted the write locally, but it was misleading because no replicas were alive to back up the data, risking permanent loss if that leader crashed.

### What would happen with `min.insync.replicas=1` in step 4?
An acks=all producer would continue to succeed even with two brokers down because a minimum requirement of one in-sync replica is satisfied by the leader alone.

### Which `acks` setting would you choose for the fraud transactions of *Card Sentinel* (Exercise 6), and which for website click events? Why?
I would choose acks=all for fraud transactions to guarantee zero data loss.

And acks=0 or acks=1 for website clicks where high throughput matters more than losing some click data.

### Why are the numbers of step 6 optimistic compared with a real cluster spread over several machines or availability zones?
The step 6 numbers are optimistic because all brokers and the producer run locally on the same machine via Docker, avoiding the network latency and physical switch overhead of a real environment.