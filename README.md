# Redify (Any database as redis)

[![Testing Status](https://github.com/demdxx/redify/workflows/Tests/badge.svg)](https://github.com/demdxx/redify/actions?workflow=Tests)
[![Publish Docker Status](https://github.com/demdxx/redify/workflows/Publish/badge.svg)](https://github.com/demdxx/redify/actions?workflow=Publish)
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
APP_BUILD_TAGS=pgx,clickhouse,mysql,mssql,kafka,redispub,nats make build
```

```sh
redify --conf docker/example.config.yml
```

## Run in docker

```sh
docker run -v ./my.config.yml:/config.yml -it --rm demdxx/refidy:latest --conf /config.yml
```

## Config example

```yml
cache:
  # redis://host:port/{dbnum}?max_retries=0&min_retry_backoff=10s&max_retry_backoff=10s&dial_timeout=3s&read_timeout=3s&write_timeout=3s&pool_fifo=false&pool_size=10&min_idle_conns=60s&max_conn_age=60s&pool_timeout=300s&idle=100s&idle_check_frequency=3s&ttl=200s
  connect: "memory"
  size: 1000 # Max capacity
  ttl: 60s # Seconds
sources:
  - connect: "postgres://dbuser:password@pgdb:5432/project?sslmode=disable"
    # Predefined in the postgresql notification channel
    notify_channel: redify_update
    binds:
    - dbnum: 0
      key: "post_{{slug}}"
      get_query: "SELECT * FROM posts WHERE slug = {{slug}} AND deleted_at IS NULL LIMIT 1"
      list_query: "SELECT slug FROM posts WHERE deleted_at IS NULL"
    - dbnum: 1
      # Automaticaly prepare requests for table `users` with key field `username`
      key: "user_{{username}}"
      table_name: "users"
      readonly: yes
    - dbnum: 2
      key: "document_{{type}}_{{slug}}"
      get_query: |
        SELECT * FROM documents WHERE type={{type}} AND slug={{slug}} AND deleted_at IS NULL LIMIT 1
      list_query: |
        SELECT type, slug FROM documents WHERE deleted_at IS NULL
      upsert_query: |
        INSERT INTO documents (slug, type, title, content)
          VALUES ({{slug}},{{type}},{{title}},{{content}})
          ON CONFLICT (slug, type) DO UPDATE SET title={{title}}, content={{content}}, deleted_at=NULL
      del_query: |
        UPDATE documents SET deleted_at=NOW() WHERE type={{type}} AND slug={{slug}}
  - connect: "mysql://dbuser:password@mysql:3306/project"
    binds:
    - dbnum: 0
      key: "wp_post_{{slug}}"
      get_query: "SELECT * FROM wp_posts WHERE slug = {{slug}} AND deleted_at IS NULL LIMIT 1"
      list_query: "SELECT slug FROM wp_posts WHERE deleted_at IS NULL"
  - connect: nats://nats:4442/group?topics=news
    binds:
    - dbnum: 0
      key: news_notify
```

## Redis using example

For using of Redis protocol.

```sh
export SERVER_REDIS_LISTEN=:8080
export SERVER_REDIS_READ_TIMEOUT=120s
```

```sh
> redis-cli -p 8081 -h hostname
hostname:8081> keys *o*
1) "post_hello"
2) "post_bye"
3) "document_docx_main"
4) "document_pdf_help"
hostname:8081> get post_bye
"{\"content\":\"Bye everyone\",\"created_at\":\"2021-11-06T20:03:56.218629Z\",\"deleted_at\":null,\"id\":4,\"slug\":\"bye\",\"title\":\"Bye world\",\"updated_at\":\"2021-11-06T20:03:56.218629Z\"}"
hostname:8081> hgetall post_hello
 1) "content"
 2) "Hello everyone"
 3) "created_at"
 4) "2021-11-06T20:03:56.218629Z"
 5) "deleted_at"
 6) "null"
 7) "id"
 8) "3"
 9) "slug"
10) "hello"
11) "title"
12) "Hello world"
13) "updated_at"
14) "2021-11-06T20:03:56.218629Z"
hostname:8081> hget post_hello title
"Hello world"
```

## HTTP example

For using of HTTP protocol.

```sh
export SERVER_HTTP_LISTEN=:8080
export SERVER_HTTP_READ_TIMEOUT=120s
```

> GET /:dbnum/:key

```sh
curl -XGET "http://localhost:8080/0/post_hello"
{"status":"OK", "result":{"content":"Hello everyone","created_at":"2021-11-06T20:03:56.218629Z","deleted_at":null,"id":3,"slug":"hello","title":"Hello world","updated_at":"2021-11-06T20:03:56.218629Z"}}
```

> PUT /:dbnum/:key
> POST /:dbnum/:key

```sh
curl -d "@data.json" -XPOST "http://localhost:8080/0/list/post_hello"
{"status":"OK"}
```

> GET /:dbnum/list/:pattern

```sh
curl -XGET "http://localhost:8080/0/list/post_*"
{"status":"OK", "result":["post_post-1","post_post-2","post_hello","post_bye"]}
```

> DELETE /:dbnum/:key

```sh
curl -XDELETE "http://localhost:8080/0/post_hello"
{"status":"OK"}
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

## Event streaming

In some cases, it will be convenient to use storage keys as a mechanism
for publishing events to message publishing systems (queues).
It can be used to publish intrasystem events or form pending actions.

### Build
```ss
APP_BUILD_TAGS=kafka,redispub,nats make build
```

### Config
```yaml
sources:
  # (kafka|nats|redispub)://hostname:port/group_name?topics=name_in_stream
  - connect: nats://nats:4442/group?topics=news
    binds:
    - dbnum: 0
      key: news_notify
```

### Using
```sh
hostname:8081> set news_notify '{"action":"view","id":100,"ts":199991229912}'
OK
```

## Support Redis commands

* SELECT \[dbnum\]
* KEYS \[pattern\]
* GET key
* MGET key1 key2 ... keyN
* HGET key fieldname
* HGETALL key
* SET key value
* MSET key1 value1 key2 value2 ... keyN valueN
* PING
* QUIT

## TODO

* [X] PGX PostgreSQL driver support
* [X] MySQL driver support
* [X] Sqlite driver support
* [X] MSSQL driver support
* [X] Oracle driver support
* [X] Stream Publishing driver (Kafka,NATS,Redis Pub)
* [X] Clickhouse driver support
* [ ] Add personal cache to every bind separately
* [ ] Cassandra driver support
* [ ] MongoDB driver support
* [ ] NextJS application example
