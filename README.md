# Fingo

Finance tracking app for learning Go.

* **Templating:** [Templ](https://templ.guide/) – Type-safe HTML components rendered directly in Go.
* **Database Migrations:** [Goose](https://github.com/pressly/goose) – Reliable, version-controlled schema management.
* **Database Access:** [sqlc](https://sqlc.dev/) – Generates type-safe, idiomatic Go code from raw SQL queries.


## Database migration

Goose is used for db migration.
To run a migration:

- cd into schema dir
```shell
cd sql/schema
```

- run goose command

```shell
goose postgres "postgres://<user>:<password>@<host>:<port>/<db>" <up | down>
```

example:
```shell
goose postgres "postgres://user:pass@localhost:5432/fingo" up
```

## SQL

To add SQL query, add it to sql/queries. Normally one table has one sql file.
Afterwards, run
```shell
sqlc generate
```
and you new query will be traslated to Go and put at internal/database module.

## Templ

This project uses Templ for html templating. Run `templ generate` after updating .templ files.

## Tests debuging

You can start debginging session with:
```shell
dlv test --headless -l 127.0.0.1:2345
```

and can use this launch.json:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Connect to Remote Server",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "127.0.0.1",
            "remotePath": "${workspaceFolder}"
        }
    ]
}
```
