package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bravetools/bravetools/shared"

	// import sqlite driver
	_ "modernc.org/sqlite"
)

// OpenDB opens database
func OpenDB(filepath string) (db *sql.DB, err error) {
	//log.Println("Connecting to SQlite database " + filepath)

	if !shared.FileExists(filepath) {
		return nil, fmt.Errorf("Database file %s not present", filepath)
	}

	db, err = sql.Open("sqlite", filepath)
	return db, err
}

// InitDB creates an empty database
func InitDB(filepath string) error {

	//os.Remove(filepath)
	//log.Println("Creating database. ", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	log.Println("Database file created")

	db, err := OpenDB(filepath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	log.Println("Creating units table ..")
	sqlStatement := `CREATE TABLE units (
		"id" integer NOT NULL PRIMARY KEY,
		"uid" TEXT(50) NOT NULL,
		"name" TEXT(50) NOT NULL COLLATE NOCASE,
		"date" TEXT(50) NOT NULL,
		"data" BLOB
	);
	
	CREATE INDEX uid_IDX ON units (uid);`

	statement, err := db.Prepare(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// InsertUnitDB inserts a unit into database
func InsertUnitDB(db *sql.DB, unit BraveUnit) (int64, error) {
	defer db.Close()

	// TODO: duplicate unit names could exist in DB. If unit name required to be unique it should be checked earlier.

	//log.Println("Inserting unit ..")
	insertUnit := `INSERT INTO units(uid, 
									name,
									date,
									data)
									VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertUnit)
	if err != nil {
		return 0, errors.New("Failed to prepare SQL statement " + err.Error())
	}

	r, err := statement.Exec(unit.UID,
		unit.Name,
		unit.Date,
		unit.Data)
	if err != nil {
		return 0, errors.New("Failed to execute SQL statement " + err.Error())
	}

	//log.Printf("Unit inserted. Unit name: %v", unit.Name)

	id, _ := r.LastInsertId()

	return id, nil
}

// DeleteUnitDB deletes a unit from database
func DeleteUnitDB(db *sql.DB, name string) error {
	defer db.Close()
	//log.Println("Deleting unit ...")
	var sql = `DELETE FROM units WHERE name=?;`
	statement, err := db.Prepare(sql)
	if err != nil {
		return errors.New("Error preparing SQL: " + err.Error())
	}

	res, err := statement.Exec(name)
	if err != nil {
		return errors.New("Error deleting unit: " + err.Error())
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("No records to delete")
	}

	//log.Println("Unit deleted: ", name)
	return nil
}

// GetUnitDB returns a unit from database by name
func GetUnitDB(db *sql.DB, name string) (unit Unit, err error) {
	defer db.Close()
	unit, err = unitByName(db, name)
	if err != nil {
		return unit, err
	}

	if unit.Name == "" {
		return unit, errors.New("Unit not found")
	}

	return unit, nil
}

// GetAllUnitsDB returns all units
func GetAllUnitsDB(db *sql.DB) (units []Unit, err error) {
	defer db.Close()
	rows, err := db.Query("SELECT * FROM units")
	if err != nil {
		log.Fatal(err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var id int
		var uid string
		var name string
		var date string
		var data []byte
		rows.Scan(&id, &uid, &name, &date, &data)
		var unitData UnitData
		err = json.Unmarshal(data, &unitData)
		if err != nil {
			return units, err
		}
		unit := Unit{
			UID:  uid,
			Name: name,
			Date: date,
			Data: unitData,
		}
		units = append(units, unit)
	}

	return units, nil
}

func unitByName(db *sql.DB, name string) (unit Unit, err error) {
	sqlStatement, err := db.Prepare("SELECT * FROM units WHERE name=? COLLATE NOCASE")
	if err != nil {
		return unit, err
	}

	rows, err := sqlStatement.Query(name)
	defer rows.Close()
	for rows.Next() {
		var id int64
		var uid string
		var name string
		var date string
		var data []byte
		err = rows.Scan(&id, &uid, &name, &date, &data)
		if err != nil {
			return unit, err
		}

		var unitData UnitData
		err = json.Unmarshal(data, &unitData)
		if err != nil {
			return unit, err
		}

		unit.ID = id
		unit.UID = uid
		unit.Name = name
		unit.Date = date
		unit.Data = unitData
	}

	return unit, nil
}
