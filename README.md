# Redify (Any database as redis)

[![Build Status](https://github.com/demdxx/redify/workflows/run%20tests/badge.svg)](https://github.com/demdxx/redify/actions?workflow=run%20tests)
[![Coverage Status](https://coveralls.io/repos/github/demdxx/redify/badge.svg?branch=main)](https://coveralls.io/github/demdxx/redify?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/demdxx/redify)](https://goreportcard.com/report/github.com/demdxx/redify)
[![GoDoc](https://godoc.org/github.com/demdxx/redify?status.svg)](https://godoc.org/github.com/demdxx/redify)

> License Apache 2.0

Redify is the optimized key-value proxy for quick access and cache
of any other database throught Redis and/or HTTP protocol.
Can be used for any kind of content web project to accelerate access to data
to make it faster, simpler and more reliable.

Modern websites it's a banch of small services which use many different databases.
Almost always for data reading used some key-value storage for caches.
The project will help to create more simple content website without using connection
to R_DBMS with simple wide used protocol without implementation of complex
application controllers and custom caches.

## Build project

```sh
APP_BUILD_TAGS=pgx,mysql,mssql make build
```

## Redis using example

```sh
> redis-cli -p 8081 -h hostname
hostname:8081> keys *o*
1) "post_hello"
2) "post_bye"
3) "document_docx_main"
4) "document_pdf_help"
hostname:8081> get post_bye
"{\"content\":\"Bye everyone\",\"created_at\":\"2021-11-06T20:03:56.218629Z\",\"deleted_at\":null,\"id\":4,\"slug\":\"bye\",\"title\":\"Bye world\",\"updated_at\":\"2021-11-06T20:03:56.218629Z\"}"
```

## Cache invalidation notifications

### PostgreSQL

PostgreSQL supports `pg_notify` precedure to notify about any kind of changes in data.

```sql
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

CREATE TRIGGER products_notify_event
AFTER INSERT OR UPDATE OR DELETE ON products
    FOR EACH ROW EXECUTE PROCEDURE notify_event();
```

## Support Redis commands

* SELECT \[dbnum\]
* KEYS \[pattern\]
* GET key
* MGET key1 key2 ... keyN
* SET key value
* MSET key1 value1 key2 value2 ... keyN valueN
* PING
* QUIT

## TODO

* [x] PGX PostgreSQL driver support
* [X] MySQL driver support
* [X] Sqlite driver support
* [X] MSSQL driver support
* [ ] Oracle driver support
* [ ] Cassandra driver support
* [ ] MongoDB driver support
* [ ] NextJS application example
