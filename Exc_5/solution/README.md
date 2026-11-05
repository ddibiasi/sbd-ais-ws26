# Exercise 5: Line Guardian (Manuth)

Files in this folder: `decision_record.md`, `scoreboard_output.txt`, `code/plant.py`, `code/detector_service.py`.

## Stage A: Today's system

### Small plant (20 machines x 10 Hz = 200 readings/s)

|                                         | Readings            |
| --------------------------------------- | ------------------- |
| Rows in table `readings` for this run | 62,915              |
| Read by the detector                    | 51,055              |
| Missing (never seen by the detector)    | 11,860 (about 19 %) |

51,055 + 11,860 = 62,915, so the missing readings are in the table. The detector
kept up (backlog 0) but never saw them.

Why the detector never saw them: an id is assigned at INSERT, but a row is visible
to other sessions only after COMMIT. In the timeline, gateway 1 inserts ids 100-114
and commits 100 ms later, gateway 2 inserts ids 115-129 and commits at once. If the
detector polls in between, it sees only 115-129 and remembers last_id = 129. When
gateway 1 commits, its rows 100-114 are below last_id, and `WHERE id > last_id`
never returns them. PostgresSink in plant.py makes this visible: each gateway sleeps
0 to 120 ms before COMMIT, and the four gateways commit in a different order than
they inserted.

### Full plant (60 machines x 10 Hz = 600 readings/s)

The plant sent 49,320 readings in 83 s (about 600/s). The model needs 3 ms per
reading, so one detector processes at most about 333 readings/s. It falls behind by
about 267 readings/s, which is roughly 16,000 rows of backlog per minute.
Delay and backlog therefore grow continuously and never recover while the plant runs.
Missing keeps growing too, because of the commit-order problem above.
Observed in the detector output: 6937 : delay, backlog and missing at the start and at the end of the run.

## Stage B: Decision record

Options:

1. Keep PostgreSQL and poll it better (several pollers, each for some machines).
2. The gateways write every reading into Kafka.

Option 1 against the requirements:

- R1 at risk: 100,000 inserts/s plus polling on one database.
- R2 at risk: the id gap problem stays (ids are assigned before commit), and
  splitting machines over pollers needs extra care.
- R3 at risk: polling adds delay, and every poller adds load.
- R4 fails: every new consumer polls PostgreSQL and adds load to it.

Option 2 against the requirements:

- R1 met: more partitions allow more detector instances.
- R2 met: key = machine_id, so one machine goes to one partition, in order.
- R3 met: the detector reads events as they are written, with no polling interval.
- R4 met: every consumer has its own consumer group and offsets, and PostgreSQL is not touched.

Decision: Kafka, topic `readings`, key = machine_id. PostgreSQL stays the system of
record for the machine registry.

Numbers:

- 600 readings/s x 3 ms = 1.8 s of work per second, so at least 2 detector
  instances. I use 3 for headroom (about 0.6 s of work per second each).
- 100,000 readings/s x 3 ms = 300 s of work per second, so at least 300 instances
  (about 350 with headroom), and at least that many partitions.

Tipping point: polling a table is fine for one consumer at a low rate (up to roughly
a few hundred writes per second) without replay or ordering needs. A small office with
one reading per minute is far below that. A log wins at high rates, with several
independent consumers, or when ordering and replay matter.

## Stage C: Implementation

Topics, created before the plant started:

    kafka-topics.sh --bootstrap-server localhost:9092 --create --topic readings --partitions 6 --replication-factor 1
    kafka-topics.sh --bootstrap-server localhost:9092 --create --topic anomalies --partitions 6 --replication-factor 1

KafkaSink (`code/plant.py`):

- key = machine_id, so all readings of a machine go to one partition, in order (R2).
- enable.idempotence and acks=all: no duplicates or reordering inside a partition.
- linger.ms=20 and lz4 compression: batching, no flush per reading.
- On BufferError the sink polls and retries, so no reading is dropped.

Detector (`code/detector_service.py`):

- Topic `readings`, one shared group `detector`, 3 instances (det-1, det-2, det-3).
- auto.offset.reset=latest: a live detector must not replay readings of old plant
  runs (3 ms each would delay real alerts). This requires starting the detectors
  before the plant.
- on_revoke calls model.forget() for the machines of the lost partitions. Otherwise
  the old next_seq would produce false "missing" counts if the partition comes back.
- Anomaly events are keyed by machine_id.

Result (`scoreboard_output.txt`):

- Faults detected: 3/3, mean delay 6.9 s, false alarms 0.
- Reading to anomaly event latency: p50 0.08 s, p95 1.52 s, max 2.80 s.
- R2: 0 missing, 0 out of order on all three instances. Duplicate anomaly events: 0.
- All measurable checks PASS.

Observations:

- The detection delay (6 to 7 s) is longer than the latency (p95 1.5 s) because the
  model needs a window of 50 readings (5 s at 10 Hz) before the drifting vibration
  pushes the score over the threshold. R3 measures the time after the reading.
- The detectors joined one after another (det-1, det-2, det-3), and each join
  caused a rebalance (det-1: 6 partitions, then 3, then 2). When I stopped det-1 at
  the end, its partitions moved to det-2 and det-3. missing stayed at 0 on all instances.
- Not measured: R1 at 100,000 readings/s. It would need about 300 partitions and
  instances.
