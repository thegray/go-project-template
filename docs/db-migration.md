# Database Migration Process
 
> Plain SQL files as source of truth. Developers own the files, DBAs own execution.

---

## Overview
 
All schema changes are written as plain `.sql` migration files, versioned alongside the application code. Developers create and maintain these files. DBAs review and execute them in staging and production using their own tooling.
 
The app **never runs migrations automatically** in any non-local environment.

---

## Directory Structure
 
```text
/myapp
└── migrations/
     ├── 000001_create_users_table.up.sql
     ├── 000001_create_users_table.down.sql
     ├── 000002_add_orders_table.up.sql
     ├── 000002_add_orders_table.down.sql
     └── 000003_add_index_users_email.up.sql
```

Each migration has an `up` file (applies the change) and a `down` file (rolls it back). The numeric prefix guarantees execution order.

---

## Creating a New Migration
 
Use `golang-migrate` CLI to generate consistently named files:

```bash
migrate create -ext sql -dir migrations -seq <descriptive_name>
```

Or use the repo Makefile wrapper:

```bash
make migrate-create name=create_users_table
```
 
Example:
 
```bash
migrate create -ext sql -dir migrations -seq create_users_table
# creates:
#   migrations/000001_create_users_table.up.sql
#   migrations/000001_create_users_table.down.sql
```
 
Then fill in the generated files with your SQL:
 
```sql
-- 000001_create_users_table.up.sql
CREATE TABLE users (
    id         VARCHAR(36)  NOT NULL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```
 
```sql
-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

## Local Development
 
Developers run migrations locally using `golang-migrate` CLI directly:
 
```bash
# apply all pending migrations
migrate -path migrations -database "mysql://root:root@localhost/myapp" up
 
# rollback last migration
migrate -path migrations -database "mysql://root:root@localhost/myapp" down 1
 
# check current migration version
migrate -path migrations -database "mysql://root:root@localhost/myapp" version
```

## Staging & Production
 
Developers **do not execute** migrations in staging or production. The process is:
 
1. Include migration files in the release — committed to the repo, reviewed in the PR
2. List all new migration files in the release notes or change ticket
3. Hand off `.sql` files to the DBA team
4. DBA reviews, schedules, and executes using their own tooling
5. DBA confirms execution — release can proceed
