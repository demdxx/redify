CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$

    DECLARE
        data json;
        notification json;

    BEGIN

        -- Convert the old or new row to JSON, based on the kind of action.
        -- Action = DELETE?             -> OLD row
        -- Action = INSERT or UPDATE?   -> NEW row
        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;

        -- Contruct the notification as a JSON string.
        notification = json_build_object(
                          'table',TG_TABLE_NAME,
                          'action', TG_OP,
                          'data', data);

        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('redify_update', notification::text);

        -- Result is ignored since this is an AFTER trigger
        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now(); 
   RETURN NEW;
END;
$$ language 'plpgsql';

-- POSTS

CREATE TABLE posts
( id                      BIGSERIAL                   PRIMARY KEY
, slug                    VARCHAR(128)                NOT NULL        UNIQUE

, title                   VARCHAR(512)                NOT NULL
, content                 TEXT                        NOT NULL

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON posts FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

INSERT INTO posts (slug, title, content) VALUES
  ('hello', 'Hello world', 'Hello everyone'),
  ('bye', 'Bye world', 'Bye everyone');

-- DOCUMENTS

CREATE TABLE documents
( id                      BIGSERIAL                   PRIMARY KEY
, title                   VARCHAR(512)                NOT NULL
, content                 TEXT                        NOT NULL

, slug                    VARCHAR(128)                NOT NULL        UNIQUE
, type                    VARCHAR(64)                 NOT NULL

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP

, UNIQUE (slug, type)
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON documents FOR EACH ROW EXECUTE PROCEDURE updated_at_column();

CREATE TRIGGER documents_notify_event
AFTER INSERT OR UPDATE OR DELETE ON documents
    FOR EACH ROW EXECUTE PROCEDURE notify_event();

INSERT INTO documents (slug, type, title, content) VALUES
  ('main', 'docx', 'Main', 'Secret info'),
  ('help', 'pdf', 'Onboarding', 'Greetings');

-- USERS

CREATE TABLE users
( id                      BIGSERIAL                   PRIMARY KEY
, username                VARCHAR(64)                 NOT NULL        UNIQUE
, email                   VARCHAR(64)                 NOT NULL

, created_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, updated_at              TIMESTAMP                   NOT NULL        DEFAULT NOW()
, deleted_at              TIMESTAMP
);

CREATE TRIGGER updated_at_triger BEFORE UPDATE
    ON users FOR EACH ROW EXECUTE PROCEDURE updated_at_column();
