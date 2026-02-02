# Fingo

Finance tracking app for learning Go.


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
