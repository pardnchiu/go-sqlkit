# go-sqlite - Documentation

> Back to [README](./README.md)

## Prerequisites

- Go 1.25 or higher
- CGO enabled (`go-sqlite3` requires a C compiler)

## Installation

### Using go get

```bash
go get github.com/pardnchiu/go-sqlite
```

### From Source

```bash
git clone https://github.com/pardnchiu/go-sqlite.git
cd go-sqlite
go build ./...
```

## Usage

### Basic

Create a connection and perform basic CRUD operations:

```go
package main

import (
	"fmt"
	"log"

	goSqlite "github.com/pardnchiu/go-sqlite"
	"github.com/pardnchiu/go-sqlite/core"
)

func main() {
	db, err := goSqlite.New(core.Config{
		Path: "app.db",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table
	err = db.Write.Table("users").Create(
		core.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		core.Column{Name: "name", Type: "TEXT", IsNullable: false},
		core.Column{Name: "email", Type: "TEXT", IsUnique: true},
		core.Column{Name: "age", Type: "INTEGER", Default: 0},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Insert data
	id, err := db.Write.Table("users").Insert(map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("inserted id:", id)
}
```

### Struct Binding

Map query results directly to structs via `Bind()`:

```go
type User struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

// Bind single struct
var user User
_, err := db.Read.Table("users").
	WhereEq("id", 1).
	Bind(&user).
	Get()
if err != nil {
	log.Fatal(err)
}

// Bind slice
var users []User
_, err = db.Read.Table("users").
	WhereGe("age", 18).
	OrderBy("name", core.Asc).
	Bind(&users).
	Get()
if err != nil {
	log.Fatal(err)
}
```

### Advanced Queries

```go
// Combined conditions: WHERE + OR + JOIN + GROUP BY + HAVING
var results []Result
_, err := db.Read.Table("orders").
	Select("user_id", "SUM(amount) AS total").
	Join("users", "users.id = orders.user_id").
	WhereGe("created_at", "2026-01-01").
	OrWhereEq("status", "vip").
	GroupBy("user_id").
	HavingGt("total", 1000).
	OrderBy("total", core.Desc).
	Limit(10).
	Bind(&results).
	Get()
if err != nil {
	log.Fatal(err)
}

// First / Last
var latest User
_, err = db.Read.Table("users").
	OrderBy("id", core.Desc).
	Bind(&latest).
	First()

// Count
count, err := db.Read.Table("users").
	WhereNotNull("email").
	Count()

// Paginated query with Total
var page []User
_, err = db.Read.Table("users").
	Total().
	Limit(20).
	Offset(40).
	Bind(&page).
	Get()
```

### Write Operations

```go
// Batch insert
affected, err := db.Write.Table("users").InsertBatch([]map[string]any{
	{"name": "Bob", "email": "bob@example.com", "age": 25},
	{"name": "Charlie", "email": "charlie@example.com", "age": 28},
})

// Conflict handling (Upsert)
id, err := db.Write.Table("users").
	Conflict(core.Replace).
	Insert(
		map[string]any{"name": "Alice", "email": "alice@example.com", "age": 31},
		map[string]any{"age": 31}, // ON CONFLICT DO UPDATE SET
	)

// Update
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Update(map[string]any{"age": 32})

// Increase / Decrease / Toggle
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Increase("age", 1).
	Update()

affected, err = db.Write.Table("posts").
	WhereEq("id", 5).
	Toggle("is_published").
	Update()

// Delete (requires explicit WHERE or force)
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Delete()

// Context support
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

_, err = db.Read.Table("users").
	Context(ctx).
	WhereEq("id", 1).
	Bind(&user).
	Get()

// Raw SQL
rows, err := db.Read.Raw().Query("SELECT * FROM users WHERE age > ?", 18)
result, err := db.Write.Raw().Exec("DELETE FROM sessions WHERE expired_at < ?", time.Now())
```

## API Reference

### goSqlite.New

```go
func New(c core.Config) (*core.Connector, error)
```

Create a read-write separated SQLite connection. The write side uses a single connection with WAL mode; the read side uses a configurable connection pool.

### core.Config

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Path` | `string` | Yes | SQLite database file path |
| `Lifetime` | `int` | No | Max read connection lifetime in seconds (default 120) |
| `MaxOpenConns` | `int` | No | Max open read connections (default 50) |
| `MaxIdleConns` | `int` | No | Max idle read connections (default 25) |

### core.Connector

| Method | Signature | Description |
|--------|-----------|-------------|
| `Query` | `(key, query string, args ...any) (*sql.Rows, error)` | Execute query via read connection |
| `QueryContext` | `(ctx, key, query string, args ...any) (*sql.Rows, error)` | Query with Context |
| `Exec` | `(key, query string, args ...any) (sql.Result, error)` | Execute via write connection |
| `ExecContext` | `(ctx, key, query string, args ...any) (sql.Result, error)` | Execute with Context |
| `Close` | `()` | Close both read and write connections |

### core.Builder — Table Operations

| Method | Signature | Description |
|--------|-----------|-------------|
| `Table` | `(name string) *Builder` | Set the target table |
| `Raw` | `() *sql.DB` | Get the underlying `sql.DB` |
| `Create` | `(columns ...Column) error` | Create table (CREATE TABLE IF NOT EXISTS) |

### core.Builder — Query

| Method | Signature | Description |
|--------|-----------|-------------|
| `Select` | `(columns ...string) *Builder` | Specify SELECT columns |
| `Join` | `(table, on string) *Builder` | INNER JOIN |
| `LeftJoin` | `(table, on string) *Builder` | LEFT JOIN |
| `OrderBy` | `(column string, direction ...direction) *Builder` | Sort order |
| `GroupBy` | `(columns ...string) *Builder` | Group by |
| `Limit` | `(num ...int) *Builder` | Limit rows (two args: offset, limit) |
| `Offset` | `(num int) *Builder` | Offset |
| `Total` | `() *Builder` | Enable COUNT(*) OVER() pagination |
| `Context` | `(ctx context.Context) *Builder` | Set Context |
| `Bind` | `(target any) *Builder` | Bind target struct/slice pointer |
| `Get` | `() (*sql.Rows, error)` | Execute SELECT query |
| `First` | `() (*sql.Row, error)` | Get newest row (reversed order) |
| `Last` | `() (*sql.Row, error)` | Get oldest row |
| `Count` | `() (int64, error)` | Get row count |

### core.Builder — Where Clauses

| Method | Description |
|--------|-------------|
| `Where(condition, args...)` | Custom AND condition |
| `WhereEq(column, value)` | `= ?` |
| `WhereNotEq(column, value)` | `!= ?` |
| `WhereGt(column, value)` | `> ?` |
| `WhereLt(column, value)` | `< ?` |
| `WhereGe(column, value)` | `>= ?` |
| `WhereLe(column, value)` | `<= ?` |
| `WhereIn(column, values)` | `IN (?, ...)` |
| `WhereNotIn(column, values)` | `NOT IN (?, ...)` |
| `WhereNull(column)` | `IS NULL` |
| `WhereNotNull(column)` | `IS NOT NULL` |
| `WhereBetween(column, start, end)` | `BETWEEN ? AND ?` |

All Where methods have corresponding `OrWhere` variants (joined with OR).

### core.Builder — Having Clauses

Same method signatures as the Where series, prefixed with `Having` / `OrHaving`.

### core.Builder — Write

| Method | Signature | Description |
|--------|-----------|-------------|
| `Conflict` | `(conflict conflict) *Builder` | Set conflict strategy |
| `Insert` | `(data ...map[string]any) (int64, error)` | Insert row (second map for ON CONFLICT UPDATE) |
| `InsertBatch` | `(data []map[string]any) (int64, error)` | Batch insert |
| `Increase` | `(column string, num ...int) *Builder` | Increment column |
| `Decrease` | `(column string, num ...int) *Builder` | Decrement column |
| `Toggle` | `(column string) *Builder` | Toggle boolean |
| `Update` | `(data ...map[string]any) (int64, error)` | Update rows |
| `Delete` | `(force ...bool) (int64, error)` | Delete (no WHERE requires force=true) |

### core.Column

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Column name |
| `Type` | `string` | SQL type |
| `IsPrimary` | `bool` | Primary key |
| `IsNullable` | `bool` | Allow NULL |
| `AutoIncrease` | `bool` | Auto increment |
| `IsUnique` | `bool` | Unique constraint |
| `Default` | `any` | Default value |
| `ForeignKey` | `*Foreign` | Foreign key reference |

### Conflict Strategy Constants

| Constant | Description |
|----------|-------------|
| `core.Ignore` | INSERT OR IGNORE |
| `core.Replace` | INSERT OR REPLACE |
| `core.Abort` | INSERT OR ABORT |
| `core.Fail` | INSERT OR FAIL |
| `core.Rollback` | INSERT OR ROLLBACK |

### Sort Direction Constants

| Constant | Description |
|----------|-------------|
| `core.Asc` | Ascending order |
| `core.Desc` | Descending order |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
