# Exercise 8: The crash report, live

Name: Manuth Perera

## 1. Decision

I chose architecture B: alert early, correct later. It is one stream with a 30 s watermark and output mode `update`.

- The on-call engineer gets an alert as soon as a user has 3 severe crashes in an interval, so nobody waits for late records.
- The operations team gets a correct report because late records (up to 30 s) still change the count, and a new version of the alert is written.
- The example (watermark 5 s, mode `append`) does not serve the operations team: records later than 5 s are dropped, so counts are wrong or alerts are missing.

Cost:
- An alert is not final until the watermark has passed its interval.
- The same key (interval + user) can appear several times in `crash_alerts`, as an update (new count) or as a repeat (same count, written again after the restart). Every reader must handle this: the pager keeps the first message per key, the ops view keeps the latest count per key.
- There is only one pipeline, so alerts and report cannot disagree (unlike architecture C).

## 2. The query in Spark

How it runs:
- Spark runs the query in micro-batches. Each one reads the new offsets of the topic `logs`, updates the counts per 10 second interval and user, and writes the changed rows with `foreachBatch` into `crash_alerts`.
- The counts are the state. Spark keeps them between micro-batches and saves them in the checkpoint.
- The watermark (30 s) is the maximum event time seen minus 30 s. When an interval is older than that, Spark drops its state, and later records for it would be dropped.
- The checkpoint stores the offsets, the state and which micro-batch was done.

Evidence:
- ui-runs.png: the check made two runs of `crash_alerts` with the same ID (93d5b896...) but different Run IDs. The first run failed at the simulated crash (latest batch 19), the second one continued and finished at batch 61. So the restart reused the query ID, the offsets and the state from the checkpoint, and did not start from zero.
- ui-delay.png: a micro-batch takes mostly 1 to 2 seconds (peaks up to about 5 s). The global watermark gap is about 31 to 40 s, which matches the 30 s watermark. This gap does not delay the alerts, because in update mode a row is written in the micro-batch where its count reaches 3, not after the watermark. Most pages therefore come right away (p50 0.5 s). The slower ones (p95 13.9 s) are problems that reach 3 crashes only when late records arrive after the interval has ended. A micro-batch of 1 to 2 s adds a little on top. All 628 problems were paged and the p95 stays under the 20 s limit.
- ui-state.png: the query keeps about 2000 to 2800 state rows. The sawtooth shape shows that the state grows with new intervals and is cleaned up when the watermark passes old intervals. Late rows dropped by the watermark: 0, so the query loses nothing.

The checkpoint behaves like the state file approach of Exercise 7: it remembers where the job stopped and the next run continues from there. A crash between writing to Kafka and updating the checkpoint writes a micro-batch twice (a repeat), so delivery is at-least-once.

## 3. The pager

In my run the pager turned 748 messages into 628 pages. It ignored 120 messages:
- updates (107): the same key came again with a new count, because update mode writes a new version whenever the count changes (for example 3, then 4 crashes in one interval, or a late record)
- repeats (13): the same key with the same count, written again after the restart because the crashed micro-batch was run again

Without its `if key not in paged`, the on-call engineer would be paged again for every update and every repeat, so the same problem would page several times.

## 4. Output of check_all

```
THE CHECK: 260,000 log records in 130 s, 13,020 of them late (by up to 30 s);
           your query was stopped once and restarted from its checkpoint.
THE TRUTH: 628 problems (a user with more than 2 severe crashes in one 10-second interval) in 8 intervals,
           counted over every record written, late ones included.

ON-CALL ENGINEER: a page at most 20 s after the interval ends, for 95 % of the problems
  delay of the page after the interval ends             p50 0.5 s · p95 13.9 s        MET
  problems paged                                        628 of 628 (100.0 %)
  messages in crash_alerts → pages                      748 → 628
    ignored: repeats (written again after the restart)  13
    ignored: updates (a new count, same problem)        107

OPERATIONS TEAM: a correct report, at least 99 % of the rows right, late records included
  rows correct (right problem, right count)             628 of 628 (100.0 %)          MET
    missing: a problem not in the report                0
    wrong count                                         0
    extra: in the report, but not a problem             0
  read from                                             ops_view(messages)

RESULT: on-call engineer served (p95 13.9 s, needs ≤ 20 s) · operations team served (100.0 % correct, needs ≥ 99 %)
```
