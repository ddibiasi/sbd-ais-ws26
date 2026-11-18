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
