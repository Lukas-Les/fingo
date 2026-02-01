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
