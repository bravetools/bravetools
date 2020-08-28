package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var fileAuditTable = `
CREATE TABLE units (
	"id" integer NOT NULL PRIMARY KEY,
	"uid" TEXT(50) NOT NULL,
	"name" TEXT(50) NOT NULL COLLATE NOCASE,
	"date" TEXT(50) NOT NULL,
	"data" BLOB
);

CREATE INDEX uid_IDX ON file_audit (uid);
`

var tables map[string]string

func main() {

	tables = make(map[string]string)
	tables["file_audit"] = fileAuditTable

	os.Remove("brave.db")

	log.Println("Creating brave.db...")
	file, err := os.Create("brave.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("brave.db created")

	db, err := InitDB("brave.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	for key, val := range tables {
		err = CreateTable(db, key, val)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// InitDB creates a new SQLite database in current directory
func InitDB(filePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, err
	}

	return db, nil
}

// CreateTable creates a table
func CreateTable(db *sql.DB, table string, sql string) error {

	log.Println("Create " + table + "table...")
	statement, err := db.Prepare(sql)
	if err != nil {
		return errors.New("Error creating file_audit table: " + err.Error())
	}
	statement.Exec()
	log.Println("file_audit table created")

	return nil
}

// Cleanup deletes all records from all tables
func Cleanup(db *sql.DB) error {

	log.Println("Cleaning up database tables ...")
	var sql = `
	DELETE from units;
	`
	statement, err := db.Prepare(sql)
	if err != nil {
		return errors.New("Error deleteing records: " + err.Error())
	}
	statement.Exec()
	log.Println("records deleted")

	return nil
}
