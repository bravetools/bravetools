package db

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func Test_GetAllUnits(t *testing.T) {

	db, err := OpenDB("brave_test.db")
	if err != nil {
		t.Log("Failed to open db")
	}

	units, err := GetAllUnitsDB(db)
	if err != nil {
		t.Log("Error getting all units")
		t.Log("Error: ", err)
	} else {
		t.Log("Units: ", units)
	}
}

func Test_GetUnit(t *testing.T) {

	db, err := OpenDB("brave_test.db")
	if err != nil {
		t.Log("Failed to open db")
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

	db, err := OpenDB("brave_test.db")
	if err != nil {
		t.Log("Failed to open db")
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
	err := InitDB("brave_test.db")

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

	db, err := OpenDB("brave_test.db")
	if err != nil {
		t.Log("Failed to open db")
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
