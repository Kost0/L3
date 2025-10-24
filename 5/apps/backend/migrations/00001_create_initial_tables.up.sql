CREATE TABLE events (
    uuid UUID PRIMARY KEY,
    title VARCHAR(255),
    date TIMESTAMP,
    amount_of_seats INT
);

CREATE TABLE seats (
    uuid UUID PRIMARY KEY,
    is_booked BOOLEAN,
    is_paid BOOLEAN,
    event_id UUID REFERENCES events(uuid)
);