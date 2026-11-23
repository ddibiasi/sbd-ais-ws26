
# Exercise 7: From the log to the lake

Manuth

## Decision

Architecture B (the lake is the truth). The start position is read from the
lake itself (lake_position), so data and position can never disagree. A state
file (A) can get out of sync with the data after a crash. C allows duplicates
in the lake and only hides them when reading.

## Why the example fails

It always reads the topic from the beginning (read_new(None)) and writes
straight into the lake. The second run wrote all 2,000,000 records again, so
the lake held 4,000,000 records, 2,000,000 of them duplicates.

## Explanation of the result

Each run reads its start position from the lake, writes the new records to
STAGING, and moves the finished run into the lake with one rename (publish).

- Run stops while writing: only STAGING is dirty, the lake is unchanged. The
  checker shows 100,000 records not loaded yet, and the next run loads them.
- Run stops after writing but before publish: the records are not in the lake
  yet, so the next run reads them again. Same result: 100,000 not loaded, then
  loaded by the next run.
- publish is a single step, so a run is either fully in the lake or not at all.
  No record is missing and none is copied twice in any of the five situations.

## Output of check_all

```
situation                                 in lake  duplicates  missing  not loaded    run  check
-----------------------------------------------------------------------------------------------
1 first run                             2,000,000           0        0           0 106.1s  PASS
2 new records arrive, run again         2,100,000           0        0           0  18.9s  PASS
3 nothing new, run again                2,100,000           0        0           0   0.9s  PASS
4 run stops WHILE writing: right after  2,100,000           0        0     100,000         PASS
  ... then the next run                 2,200,000           0        0           0  17.2s  PASS
5 run stops AFTER writing: right after  2,200,000           0        0     100,000         PASS
  ... then the next run                 2,300,000           0        0           0  18.5s  PASS
-----------------------------------------------------------------------------------------------
Exactly once after every run: YES
```
