\## 📐 Architecture & Design (Why it looks the way it does)

| Layer                | Key Decisions                                                                                                                             | Meets Assignment Items                      |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| **Transport**        | Echo v4 + custom JWT middleware.<br>Swagger comments generate the OpenAPI spec automatically (`swag init`).                               | *“API endpoints + security middleware”*     |
| **Service**          | Pure Go interface; orchestrates *repository ↔ cache* inside an explicit transaction, then updates Redis (*cache‑aside*).                  | *Proper concurrency, clean code separation* |
| **Repository**       | **pgx + sqlc** → type‑safe, zero ORM overhead.<br>`internal/repository/pg` hides SQL; interfaces live in `internal/repository`.           | *“PostgreSQL schema & query optimisation”*  |
| **Cache**            | **go‑redis** helper; 5‑minute TTL, cache‑aside; invalidated by writers.                                                                   | *“Redis caching strategy”*                  |
| **Stream**           | Two goroutines: `Generator` (producer) and `Start` (consumer). They communicate via a buffered channel to keep back‑pressure predictable. | *“Use of Go’s concurrency model”*           |
| **Observability**    | `slog` structured logs, Echo request‑logger, optional Prometheus middleware.                                                              | *“Logging & Observability”*                 |
| **CI Tests**         | pgxmock + miniredis + table‑driven tests give deterministic unit coverage without real infra.                                             | *“Unit testing with table‑driven approach”* |
| **Containerisation** | Multi‑stage `Dockerfile` (distroless) + one‑shot `migrate` job in `docker‑compose.yml`.                                                   | *“Setup instructions (docker‑compose up)”*  |

<small>See the **diagram in `docs/architecture.svg`** for a visual of data‑flow from HTTP → service → DB|Redis → client.</small>

---

\## 🚀 Quick Start (Setup & Running Locally)

```bash
# 0️⃣ clone and prepare env‑file
git clone <your‑repo> && cd fleet-tracker
cp .env.example .env            # edit values as you like

# 1️⃣ spin up everything
docker compose up -d --build    # Postgres, Redis, migrations, API

# 2️⃣ tail logs – you should see “stream update …”
docker compose logs -f api

# 3️⃣ grab a JWT for testing
TOKEN=$(make token)             # or: go run ./cmd/token -sub dev

# 4️⃣ call the API
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/vehicle/status?id=d9c1b442-fb2f-412a-9d2a-a3ab499cd91c" | jq
```

| Task                     | Command                                                       |
| ------------------------ | ------------------------------------------------------------- |
| **Run unit tests**       | `go test ./... -race -cover`                                  |
| **Generate Swagger**     | `swag init` ➜ open `http://localhost:8080/swagger/index.html` |
| **Shut everything down** | `docker compose down -v`                                      |

---

\## 🗄️ Database Schema & Index Rationale

```sql
-- vehicle table
CREATE TABLE vehicle (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plate_number TEXT UNIQUE NOT NULL,
    last_status  JSONB NOT NULL
);

-- trips table
CREATE TABLE trips (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicle(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ,
    mileage    DOUBLE PRECISION,
    avg_speed  DOUBLE PRECISION
);

-- ⬇️ this multi‑column index is critical for the “past 24 h” query
CREATE INDEX idx_trips_vehicle_time
          ON trips (vehicle_id, start_time DESC);
```

\### EXPLAIN ANALYZE proof

> **Insert the screenshot or copy‑paste here** after you run the query below on a dataset with \~500 k rows.

```sql
EXPLAIN ANALYZE
SELECT *
  FROM trips
 WHERE vehicle_id = 'd9c1b442-fb2f-412a-9d2a-a3ab499cd91c'
   AND start_time >= NOW() - INTERVAL '24 hours'
 ORDER BY start_time DESC;
```

<details>
<summary>Sample output (text)</summary>

```
Index Scan using idx_trips_vehicle_time on trips  (cost=0.43..12.88 rows=3 width=65) 
  Index Cond: ((vehicle_id = 'd9c1b442-fb2f-412a-9d2a-a3ab499cd91c'::uuid)
               AND (start_time >= (now() - '24:00:00'::interval)))
Planning Time: 0.25 ms
Execution Time: 0.13 ms
```

</details>

Add a PNG‑or‑SVG capture of the full **`EXPLAIN ANALYZE`** in
`docs/img/explain_trips_24h.png`, then embed like:

```markdown
![Query plan for past‑24 h trips](docs/img/explain_trips_24h.png)
```

---

\## 🛡️ Security – JWT Flow for Local Dev

1. **Mint token**

   ```bash
   go run ./cmd/token -sub alice -exp 24h  # uses $JWT_SIGN_KEY
   ```
2. **Send with requests**
   `Authorization: Bearer <token>`
3. **Middleware** validates HS256 signature and puts claims in `echo.Context` under key `claims`.

---

\## 🧪 Testing Strategy

* **Repositories** — pgxmock (DB) & table‑driven cases.
* **Cache** — miniredis, TTL fast‑forward.
* **Stream** — fake service that counts `Ingest` calls under `context.WithTimeout`.

Run everything with plain `go test`; no Docker or external services required → 💯 portability.

---

\## 🌐 Sample cURL Collection

| Purpose               | Command                                                                                                                                                                                                                                      |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| *Get current status*  | `curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/vehicle/status?id=<uuid>"`                                                                                                                                                |
| *Trip history (24 h)* | `curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/vehicle/trips?id=<uuid>"`                                                                                                                                                 |
| *Manual ingest*       | `curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"vehicle_id":"<uuid>","status":{"location":[55.29,25.27],"speed":42,"timestamp":"2025-06-25T12:00:00Z"}}' http://localhost:8080/api/vehicle/ingest` |

Import the Postman collection in `docs/postman_collection.json` for easier exploration.

---

\## 📋 Design Trade‑offs & Future Work

* **Why cache‑aside?**
  Simpler failure mode: if Redis is down, the API degrades gracefully by hitting Postgres.
* **Single vehicle only (assignment scope)** — extend by sharding trips by `vehicle_id` or using time‑series DB.
* **Observability** — metrics endpoint ready; add Grafana dashboard in `docker-compose.override.yml`.
* **CI/CD** — `.github/workflows/ci.yml` builds image, runs unit tests, and pushes to GH Container Registry.

---

\## 🙌 Contributing

1. Fork → create feature branch → PR.
2. `make test && make lint` must pass.
3. Two approving reviews trigger auto‑merge.

---
