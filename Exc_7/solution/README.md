# Exercise 7
Marco Hartl

## Part 1

### Why does the checker fail?

```
Exactly once: PASS
{'topic': 2000000,
 'lake': 2000000,
 'dup': 0,
 'miss': 0,
 'lag': 0,
 'files': 4,
 'mb': 111.445519,
 'ok': True,
 'problems': []}
```


```
Exactly once: FAIL — 2,000,000 duplicates
{'topic': 2000000,
 'lake': 4000000,
 'dup': 2000000,
 'miss': 0,
 'lag': 0,
 'files': 8,
 'mb': 222.891038,
 'ok': False,
 'problems': ['2,000,000 duplicates']}
 ```

 The checker checks if there is missing data or if there are duplicates. The checker fails because in the second run there are 2 million duplicates.


### Choose an architecture

**State File (A)**
It is very simple and easy to implement. But if the process crashes after writing to lake and before saving the state to the text file, there will be a duplicate at restart because the state in the text file is one behind.

**The Lake is the Truth (B)**
It always reflects the current state as we are reading the current state, so we have no risk of duplicates or lost data.

**Remove Duplicates While Reading (C)**
The duplicates need to be stored -> costs memory. And it also can't be good to have the duplicate and leave it to the reader to deal with it.

So the best option would be B-The Lake is the Truth, as we have no risk of duplicates and we also don't have duplicates flying around.

### Code

```python
# Using Architecture B-The Lake is the Truth
def run_once():
    # 1. where did the last run stop?
    start = lake_position()
    # 2. read what is new
    new = read_new(start)
    # 3. write it
    run = write_files(new, STAGING)
    publish(run)
    
    # 4. remember where you stopped (if your architecture needs it)
    # Doesn't need it
```


### Explanation of the results
Initialy the it was written to the lake immediately, creating duplicates and causing the checker to fail.

With Architecture B, each run only reads records after lake_position() writes them to staging, and publishes them to the lake, ensuring every record appears once with no missing or duplicate records even when a run stops halfway.


### Output of check_all()

```
reset: lake, staging and state deleted

situation                                 in lake  duplicates  missing  not loaded    run  check
-----------------------------------------------------------------------------------------------
1 first run                             2,200,000           0        0           0  27.4s  PASS
2 new records arrive, run again         2,300,000           0        0           0   4.5s  PASS
3 nothing new, run again                2,300,000           0        0           0   0.1s  PASS
4 run stops WHILE writing: right after  2,300,000           0        0     100,000         PASS
  ... then the next run                 2,400,000           0        0           0   4.6s  PASS
5 run stops AFTER writing: right after  2,400,000           0        0     100,000         PASS
  ... then the next run                 2,500,000           0        0           0   4.6s  PASS
-----------------------------------------------------------------------------------------------
Exactly once after every run: YES
[('1 first run',
  {'topic': 2200000,
   'lake': 2200000,
   'dup': 0,
   'miss': 0,
   'lag': 0,
   'files': 4,
   'mb': 114.903278,
   'ok': True,
   'problems': []},
  27.398333311080933),
 ('2 new records arrive, run again',
  {'topic': 2300000,
   'lake': 2300000,
   'dup': 0,
   'miss': 0,
   'lag': 0,
   'files': 8,
   'mb': 116.059515,
   'ok': True,
   'problems': []},
  4.52270770072937),
 ('3 nothing new, run again',
  {'topic': 2300000,
   'lake': 2300000,
   'dup': 0,
   'miss': 0,
   'lag': 0,
   'files': 8,
   'mb': 116.059515,
   'ok': True,
   'problems': []},
  0.08124971389770508),
 ('4 run stops WHILE writing: right after',
  {'topic': 2400000,
   'lake': 2300000,
   'dup': 0,
   'miss': 0,
   'lag': 100000,
   'files': 8,
   'mb': 116.059515,
   'ok': True,
   'problems': []},
  None),
 ('  ... then the next run',
  {'topic': 2400000,
   'lake': 2400000,
   'dup': 0,
   'miss': 0,
   'lag': 0,
   'files': 12,
   'mb': 117.210484,
   'ok': True,
   'problems': []},
  4.5626137256622314),
 ('5 run stops AFTER writing: right after',
  {'topic': 2500000,
   'lake': 2400000,
   'dup': 0,
   'miss': 0,
   'lag': 100000,
   'files': 12,
   'mb': 117.210484,
   'ok': True,
   'problems': []},
  None),
 ('  ... then the next run',
  {'topic': 2500000,
   'lake': 2500000,
   'dup': 0,
   'miss': 0,
   'lag': 0,
   'files': 16,
   'mb': 118.366691,
   'ok': True,
   'problems': []},
  4.572195291519165)]
```