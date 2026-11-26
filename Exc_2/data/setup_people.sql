\timing on
DROP TABLE IF EXISTS people_big;
CREATE TABLE people_big (
  id SERIAL PRIMARY KEY,
  first_name TEXT,
  last_name TEXT,
  gender TEXT,
  department TEXT,
  salary INTEGER,
  country TEXT
);
\COPY people_big(first_name,last_name,gender,department,salary,country) FROM '/data/people_1M.csv' DELIMITER ',' CSV HEADER
SELECT COUNT(*) FROM people_big;
