module mecc/app

go 1.24.0

require mecc/mysqlite v0.0.0-00010101000000-000000000000

require github.com/mattn/go-sqlite3 v1.14.24 // indirect

replace mecc/mysqlite => ../mysqlite
