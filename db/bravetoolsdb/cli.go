package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	_ "modernc.org/sqlite"
)

var unitsTable = `
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

var dbpath = flag.String("p", "", "database path")
var del = flag.Bool("d", false, "delete database if exists")
var cleanup = flag.Bool("c", false, "delete all records from existing database")

func main() {
	tables = make(map[string]string)
	tables["units"] = unitsTable
	var p string

	flag.Parse()

	if *dbpath == "" {
		p = "bravetools.db"
	} else {
		p = path.Join(*dbpath, "bravetools.db")
	}

	if *del {
		err := os.Remove(p)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if *cleanup {
		db, err := InitDB(p)
		if err != nil {
			log.Fatal(err)
		}
		err = Cleanup(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Database cleaned up")
		return
	}

	log.Println("Creating bravetools.db...")
	log.Println("Database path: ", p)
	file, err := os.Create(p)
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("bravetools.db created")

	db, err := InitDB(p)
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
	db, err := sql.Open("sqlite", filePath)
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

	log.Println("Create " + table + " table...")
	statement, err := db.Prepare(sql)
	if err != nil {
		return errors.New("Error creating units table: " + err.Error())
	}
	statement.Exec()
	log.Println("units table created")

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
