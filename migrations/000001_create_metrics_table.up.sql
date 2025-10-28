CREATE TABLE metrics (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(16) NOT NULL CHECK (type IN ('gauge', 'counter')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255),
    CHECK (
        (type = 'counter' AND delta IS NOT NULL AND value IS NULL)
        OR
        (type = 'gauge' AND value IS NOT NULL AND delta IS NULL)
    )
);
