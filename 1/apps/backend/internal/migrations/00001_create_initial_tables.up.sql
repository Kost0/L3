CREATE TABLE delivery (
    id SERIAL PRIMARY KEY,
    status VARCHAR(20),
    text TEXT NOT NULL,
    send_at TIMESTAMP
);
