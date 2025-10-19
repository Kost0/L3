CREATE TABLE photos (
    uuid UUID PRIMARY KEY,
    status VARCHAR(50),
    resize_to VARCHAR(20),
    watermark_text VARCHAR(200),
    gen_thumbnail BOOLEAN
);
