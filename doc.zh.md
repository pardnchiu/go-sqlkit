# go-sqlite - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.25 或更高版本
- CGO 啟用（`go-sqlite3` 需要 C 編譯器）

## 安裝

### 使用 go get

```bash
go get github.com/pardnchiu/go-sqlite
```

### 從原始碼建置

```bash
git clone https://github.com/pardnchiu/go-sqlite.git
cd go-sqlite
go build ./...
```

## 使用方式

### 基礎用法

建立連線並執行基本 CRUD 操作：

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

	// 建立表
	err = db.Write.Table("users").Create(
		core.Column{Name: "id", Type: "INTEGER", IsPrimary: true, AutoIncrease: true},
		core.Column{Name: "name", Type: "TEXT", IsNullable: false},
		core.Column{Name: "email", Type: "TEXT", IsUnique: true},
		core.Column{Name: "age", Type: "INTEGER", Default: 0},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 插入資料
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

### Struct 綁定查詢

透過 `Bind()` 將查詢結果直接映射至 struct：

```go
type User struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Age   int    `db:"age"`
}

// 綁定單筆 struct
var user User
_, err := db.Read.Table("users").
	WhereEq("id", 1).
	Bind(&user).
	Get()
if err != nil {
	log.Fatal(err)
}

// 綁定 slice
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

### 進階查詢

```go
// 條件組合：WHERE + OR + JOIN + GROUP BY + HAVING
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

// 帶 Total 的分頁查詢
var page []User
_, err = db.Read.Table("users").
	Total().
	Limit(20).
	Offset(40).
	Bind(&page).
	Get()
```

### 寫入操作

```go
// 批次插入
affected, err := db.Write.Table("users").InsertBatch([]map[string]any{
	{"name": "Bob", "email": "bob@example.com", "age": 25},
	{"name": "Charlie", "email": "charlie@example.com", "age": 28},
})

// 衝突處理（Upsert）
id, err := db.Write.Table("users").
	Conflict(core.Replace).
	Insert(
		map[string]any{"name": "Alice", "email": "alice@example.com", "age": 31},
		map[string]any{"age": 31}, // ON CONFLICT DO UPDATE SET
	)

// 更新
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Update(map[string]any{"age": 32})

// 遞增 / 遞減 / 切換
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Increase("age", 1).
	Update()

affected, err = db.Write.Table("posts").
	WhereEq("id", 5).
	Toggle("is_published").
	Update()

// 刪除（需明確條件或 force）
affected, err = db.Write.Table("users").
	WhereEq("id", 1).
	Delete()

// Context 支援
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

_, err = db.Read.Table("users").
	Context(ctx).
	WhereEq("id", 1).
	Bind(&user).
	Get()

// 原始 SQL
rows, err := db.Read.Raw().Query("SELECT * FROM users WHERE age > ?", 18)
result, err := db.Write.Raw().Exec("DELETE FROM sessions WHERE expired_at < ?", time.Now())
```

## API 參考

### goSqlite.New

```go
func New(c core.Config) (*core.Connector, error)
```

建立讀寫分離的 SQLite 連線。寫入端為單一連線搭配 WAL 模式，讀取端為可配置的連線池。

### core.Config

| 欄位 | 型別 | 必要 | 說明 |
|------|------|------|------|
| `Path` | `string` | 是 | SQLite 資料庫檔案路徑 |
| `Lifetime` | `int` | 否 | 讀取連線最大存活秒數（預設 120） |
| `MaxOpenConns` | `int` | 否 | 讀取連線池最大開啟數（預設 50） |
| `MaxIdleConns` | `int` | 否 | 讀取連線池最大閒置數（預設 25） |

### core.Connector

| 方法 | 簽章 | 說明 |
|------|------|------|
| `Query` | `(key, query string, args ...any) (*sql.Rows, error)` | 透過讀取連線執行查詢 |
| `QueryContext` | `(ctx, key, query string, args ...any) (*sql.Rows, error)` | 帶 Context 的查詢 |
| `Exec` | `(key, query string, args ...any) (sql.Result, error)` | 透過寫入連線執行 |
| `ExecContext` | `(ctx, key, query string, args ...any) (sql.Result, error)` | 帶 Context 的執行 |
| `Close` | `()` | 關閉讀取與寫入連線 |

### core.Builder — 表操作

| 方法 | 簽章 | 說明 |
|------|------|------|
| `Table` | `(name string) *Builder` | 指定操作的表名 |
| `Raw` | `() *sql.DB` | 取得底層 `sql.DB` |
| `Create` | `(columns ...Column) error` | 建立表（CREATE TABLE IF NOT EXISTS） |

### core.Builder — 查詢

| 方法 | 簽章 | 說明 |
|------|------|------|
| `Select` | `(columns ...string) *Builder` | 指定 SELECT 欄位 |
| `Join` | `(table, on string) *Builder` | INNER JOIN |
| `LeftJoin` | `(table, on string) *Builder` | LEFT JOIN |
| `OrderBy` | `(column string, direction ...direction) *Builder` | 排序 |
| `GroupBy` | `(columns ...string) *Builder` | 分組 |
| `Limit` | `(num ...int) *Builder` | 限制筆數（雙參數為 offset, limit） |
| `Offset` | `(num int) *Builder` | 偏移量 |
| `Total` | `() *Builder` | 啟用 COUNT(*) OVER() 分頁 |
| `Context` | `(ctx context.Context) *Builder` | 設定 Context |
| `Bind` | `(target any) *Builder` | 綁定目標 struct/slice pointer |
| `Get` | `() (*sql.Rows, error)` | 執行 SELECT 查詢 |
| `First` | `() (*sql.Row, error)` | 取得最新一筆（反轉排序） |
| `Last` | `() (*sql.Row, error)` | 取得最舊一筆 |
| `Count` | `() (int64, error)` | 取得筆數 |

### core.Builder — Where 條件

| 方法 | 說明 |
|------|------|
| `Where(condition, args...)` | 自訂 AND 條件 |
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

所有 Where 方法皆有對應的 `OrWhere` 版本（使用 OR 連接）。

### core.Builder — Having 條件

與 Where 系列相同的方法簽章，前綴為 `Having` / `OrHaving`。

### core.Builder — 寫入

| 方法 | 簽章 | 說明 |
|------|------|------|
| `Conflict` | `(conflict conflict) *Builder` | 設定衝突策略 |
| `Insert` | `(data ...map[string]any) (int64, error)` | 插入單筆（第二個 map 為 ON CONFLICT UPDATE） |
| `InsertBatch` | `(data []map[string]any) (int64, error)` | 批次插入 |
| `Increase` | `(column string, num ...int) *Builder` | 欄位遞增 |
| `Decrease` | `(column string, num ...int) *Builder` | 欄位遞減 |
| `Toggle` | `(column string) *Builder` | 布林值切換 |
| `Update` | `(data ...map[string]any) (int64, error)` | 更新 |
| `Delete` | `(force ...bool) (int64, error)` | 刪除（無 WHERE 需 force=true） |

### core.Column

| 欄位 | 型別 | 說明 |
|------|------|------|
| `Name` | `string` | 欄位名稱 |
| `Type` | `string` | SQL 型別 |
| `IsPrimary` | `bool` | 主鍵 |
| `IsNullable` | `bool` | 允許 NULL |
| `AutoIncrease` | `bool` | 自動遞增 |
| `IsUnique` | `bool` | 唯一約束 |
| `Default` | `any` | 預設值 |
| `ForeignKey` | `*Foreign` | 外鍵參照 |

### 衝突策略常數

| 常數 | 說明 |
|------|------|
| `core.Ignore` | INSERT OR IGNORE |
| `core.Replace` | INSERT OR REPLACE |
| `core.Abort` | INSERT OR ABORT |
| `core.Fail` | INSERT OR FAIL |
| `core.Rollback` | INSERT OR ROLLBACK |

### 排序方向常數

| 常數 | 說明 |
|------|------|
| `core.Asc` | 升冪排序 |
| `core.Desc` | 降冪排序 |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://linkedin.com/in/pardnchiu)
