# go-expenseTracker

[Expense Tracker Rest API](https://roadmap.sh/projects/expense-tracker-api) with multi-users capability. Each user will have his own expense records.

Utilizing golang `net/http` for http server, `sqlc` and golang `sql` for database-related operation.

## Featured

- Authentication and Authorization with JWT
- Logging middleware
- Open API Spesification version 3
- Interactive API Docs

## Getting Started

> Live Demo : https://xpense.fly.biz.id/

To run in local :

- clone this repo
- create new `.env` file (see `env-sample`)
- ensure postgres instance is running. In case you don't have one, create a new postgres container

```bash
make docker-run
```

- create db tables and function, execute `schema.sql` and `functions.sql`
- run the api

```bash
make run
```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

Live reload the application:

```bash
make watch
```

Clean up binary from the last build:

```bash
make clean
```
