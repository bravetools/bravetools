package db

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testDB = "brave_test.db"

func Test_GetAllUnits(t *testing.T) {

	db, err := OpenDB(testDB)
	if err != nil {
		t.Fatal("Failed to open db: ", err)
	}

	units, err := GetAllUnitsDB(db)
	if err != nil {
		t.Log("Error getting all units")
		t.Error("Error: ", err)
	} else {
		t.Log("Units: ", units)
	}
}

func Test_GetUnit(t *testing.T) {

	db, err := OpenDB(testDB)
	if err != nil {
		t.Fatal("Failed to open db: ", err)
	}

	unit, err := GetUnitDB(db, "test")
	if err != nil {
		t.Log("Error getting unit")
		t.Log("Error: ", err)
	} else {
		t.Log("Unit ID: ", unit.ID)
		t.Log("Unit UUID: ", unit.UID)
		t.Log("Unit Name: ", unit.Name)
		t.Log("Unit Date: ", unit.Date)
		t.Log("Unit ID: ", unit.Data.CPU)
		t.Log("Unit RAM: ", unit.Data.RAM)
		t.Log("Unit IP: ", unit.Data.IP)
		t.Log("Unit Image: ", unit.Data.Image)
	}
}

func Test_DeleteUnit(t *testing.T) {

	db, err := OpenDB(testDB)
	if err != nil {
		t.Fatal("Failed to open db: ", err)
	}

	err = DeleteUnitDB(db, "test")
	if err != nil {
		t.Log("Error deleting unit")
		t.Log("Error: ", err)
	} else {
		t.Log("Unit deleted")
	}
}

func Test_InsertUnit(t *testing.T) {
	uid, _ := uuid.NewUUID()

	unitData := UnitData{
		IP:    "0.0.0.0",
		Image: "image",
		CPU:   1,
		RAM:   "2GB",
	}

	data, _ := json.Marshal(unitData)

	unit := BraveUnit{
		UID:  uid.String(),
		Name: "test3",
		Date: time.Now().String(),
		Data: data,
	}

	db, err := OpenDB(testDB)
	if err != nil {
		t.Fatal("Failed to open db: ", err)
	}

	id, err := InsertUnitDB(db, unit)
	if err != nil {
		t.Log("Error inserting unit")
		t.Log("Error: ", err)
	} else {
		t.Log("Unit inserted")
		t.Log("Record ID: ", id)
	}
}

func TestMain(m *testing.M) {
	InitDB(testDB)
	m.Run()
	err := os.Remove(testDB)
	if err != nil {
		log.Println("failed to cleanup test database file :", testDB)
	}
}
