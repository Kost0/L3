CREATE TABLE comment (
    id UUID PRIMARY KEY,
    text text,
    parent UUID DEFAULT NULL REFERENCES comment(id) ON DELETE CASCADE,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('russian', text)) STORED
);
