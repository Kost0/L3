CREATE TABLE items (
    uuid UUID PRIMARY KEY,
    title VARCHAR(100),
    price INT,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TABLE roles (
   id BIGINT PRIMARY KEY,
   name VARCHAR(100)
);

CREATE TABLE history (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    editor_id BIGINT REFERENCES roles(id),
    item_id UUID,
    operation_type VARCHAR(30),
    old_data JSONB,
    new_data JSONB,
    changed_at TIMESTAMP DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION  log_item_changes_function()
RETURNS TRIGGER AS $$
DECLARE
v_editor_role_id BIGINT;
BEGIN
    v_editor_role_id := NULLIF(current_setting('app.current_role_id', true), '')::BIGINT;

    IF (TG_OP = 'INSERT') THEN
       INSERT INTO history (editor_id, item_id, operation_type, old_data, new_data)
       VALUES (v_editor_role_id, NEW.uuid, 'INSERT', NULL, row_to_json(NEW));
    RETURN NEW;

    ELSIF (TG_OP = 'UPDATE') THEN
        IF NEW IS DISTINCT FROM OLD THEN
            INSERT INTO history (editor_id, item_id, operation_type, old_data, new_data)
            VALUES (v_editor_role_id, NEW.uuid, 'UPDATE', row_to_json(OLD), row_to_json(NEW));
    END IF;
    RETURN NEW;

    ELSIF (TG_OP = 'DELETE') THEN
        INSERT INTO history (editor_id, item_id, operation_type, old_data, new_data)
        VALUES (v_editor_role_id, OLD.uuid, 'DELETE', row_to_json(OLD), NULL);
    RETURN OLD;
    END IF;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_history_trigger
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW
EXECUTE FUNCTION log_item_changes_function();

INSERT INTO roles VALUES
  (1, 'admin'),
  (2, 'manager'),
  (3, 'viewer');