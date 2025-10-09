CREATE TABLE link (
    id UUID PRIMARY KEY,
    short_url varchar(100),
    url varchar(255)
);

CREATE TABLE link_following (
    id UUID PRIMARY KEY,
    link_id UUID REFERENCES link(id),
    time TIMESTAMP,
    user_agent VARCHAR(100),
    ip VARCHAR(50)
);
