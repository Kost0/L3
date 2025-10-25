CREATE TABLE orders (
    uuid UUID PRIMARY KEY,
    title VARCHAR(100),
    cost INT,
    items INT,
    category VARCHAR(100),
    date TIMESTAMP DEFAULT NOW()
);