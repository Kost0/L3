CREATE TABLE notify (
    id UUID PRIMARY KEY,
    status VARCHAR(20),
    text TEXT NOT NULL,
    send_at TIMESTAMP,
    email VARCHAR(100),
    tg_user VARCHAR(50)
);
