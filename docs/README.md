## Fleet Tracker Backend

A Go-based microservice for ingesting and querying vehicle info.
It contains mock streaming, JWT auth, redis caching,OpenAPI swagger, and Docker-based setup.

---
##🚀 Bringing Up the Stack

1. Build & start all containers
   docker compose up -d --build

2. Verify hello api
   curl --location 'http://127.0.0.1:8080/hello/' 

3. To view logs
   docker compose logs -f api

4. JWT token generation : Run the following to get the Bearer token for API auth  
        go run ./cmd/token -sub dev -exp 24h
   (optional) if you get an error running the above, set the JWT_SIGN_KEY using 
        export JWT_SIGN_KEY={value_from_env}

## Sample Postman APIs
1. Vehicle ingest
curl --location 'http://127.0.0.1:8080/api/vehicle/ingest' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZGl0aSIsImV4cCI6MTc1MTA4NjU4MywiaWF0IjoxNzUxMDAwMTgzfQ.hxCFDl2wULGbjisPG-yQIuG1AbeXUGx-pARKMjT829A' \
--header 'Content-Type: application/json' \
--data '{
    "vehicle_id": "d9c1b442-fb2f-412a-9d2a-a3ab499cd91c",
    "status": {
        "location": [
            55.2967,
            25.2767
        ],
        "speed": 72.5,
        "timestamp": "2025-06-26T03:11:00Z"
    }
}'

2. Get trips for given vehicle 
curl --location 'http://127.0.0.1:8080/api/vehicle/trips?id=d9c1b442-fb2f-412a-9d2a-a3ab499cd91c' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZGl0aSIsImV4cCI6MTc1MTA4NjU4MywiaWF0IjoxNzUxMDAwMTgzfQ.hxCFDl2wULGbjisPG-yQIuG1AbeXUGx-pARKMjT829A'

3. Get Vehicle status
curl --location 'http://127.0.0.1:8080/api/vehicle/status?id=d9c1b442-fb2f-412a-9d2a-a3ab499cd91c' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZGl0aSIsImV4cCI6MTc1MTA4NjU4MywiaWF0IjoxNzUxMDAwMTgzfQ.hxCFDl2wULGbjisPG-yQIuG1AbeXUGx-pARKMjT829A'

   
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


