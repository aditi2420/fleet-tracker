CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE vehicle (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plate_number  TEXT UNIQUE NOT NULL,
    last_status   JSONB NOT NULL
);

CREATE TABLE trips (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicle(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ,
    mileage    DOUBLE PRECISION DEFAULT 0,
    avg_speed  DOUBLE PRECISION DEFAULT 0
);

CREATE INDEX idx_trips_vehicle_time
          ON trips (vehicle_id, start_time DESC);
