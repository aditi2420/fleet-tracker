# Fleet Tracker Backend

A Go-based microservice for ingesting and querying vehicle info.
It contains mock streaming, JWT auth, redis caching,OpenAPI swagger, and Docker-based setup.

---
## Indexing & Performance
EXPLAIN ANALYZE
SELECT *
FROM trips
WHERE vehicle_id = '78c4ee9a-d42d-41e3-9871-23ad10bd6cc9'
AND start_time >= NOW() - INTERVAL '24 hours'
ORDER BY start_time DESC;

|QUERY PLAN                                                                                                                                |
|------------------------------------------------------------------------------------------------------------------------------------------|
|Sort  (cost=16.39..16.42 rows=12 width=64) (actual time=0.028..0.030 rows=11 loops=1)                                                     |
|  Sort Key: start_time DESC                                                                                                               |
|  Sort Method: quicksort  Memory: 25kB                                                                                                    |
|  ->  Bitmap Heap Scan on trips  (cost=4.41..16.18 rows=12 width=64) (actual time=0.020..0.021 rows=11 loops=1)                           |
|        Recheck Cond: ((vehicle_id = '78c4ee9a-d42d-41e3-9871-23ad10bd6cc9'::uuid) AND (start_time >= (now() - '24:00:00'::interval)))    |
|        Heap Blocks: exact=1                                                                                                              |
|        ->  Bitmap Index Scan on idx_trips_vehicle_time  (cost=0.00..4.40 rows=12 width=0) (actual time=0.015..0.015 rows=11 loops=1)     |
|              Index Cond: ((vehicle_id = '78c4ee9a-d42d-41e3-9871-23ad10bd6cc9'::uuid) AND (start_time >= (now() - '24:00:00'::interval)))|
|Planning Time: 0.264 ms                                                                                                                   |
|Execution Time: 0.055 ms                                                                                                                  |


