package sqlite

import (
	"errors"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/databases"
	"github.com/bkasin/gogios/helpers"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Sqlite only requires a path to the database file
type Sqlite struct {
	DBFile string `toml:"db_file"`
}

var sampleConfig = `
  ## Sqlite3 db path
  db_file = "/var/lib/gogios/gogios.db"
`

// SampleConfig returns the default config for Sqlite
func (s *Sqlite) SampleConfig() string {
	return sampleConfig
}

// Description returns a brief explanation of the database
func (s *Sqlite) Description() string {
	return "Output check data to a Sqlite3 database file"
}

// AddRow determines whether a record for the check exists in the database
// and either inserts a new row or updates the existing one
func (s *Sqlite) AddRow(check gogios.Check) error {
	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if db.NewRecord(check) {
		db.Create(&check)
	} else {
		db.Model(check).Updates(&check)
	}

	return nil
}

// DeleteRow will remove a row from the check table based on the ID
func (s *Sqlite) DeleteRow(check gogios.Check, field string) error {
	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return err
	}
	defer db.Close()

	switch search := field; search {
	case "title":
		db.Where("title = ?", check.Title).Delete(&check)
	case "id":
		db.Delete(&check)
	default:
		err = errors.New("field needs to be title or id")
		return err
	}

	return nil
}

// GetRow returns a single row. Searches using field (title or id) and returns
// the last record that matches
func (s *Sqlite) GetRow(check gogios.Check, field string) (gogios.Check, error) {
	lastRow := gogios.Check{}

	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return lastRow, err
	}
	defer db.Close()

	if !db.HasTable(&gogios.User{}) {
		db.AutoMigrate(&gogios.User{}, &gogios.Check{})
	}

	switch search := field; search {
	case "title":
		db.Where("title = ?", check.Title).Last(&lastRow)
	case "id":
		db.Last(&lastRow)
	default:
		err = errors.New("field needs to be title or id")
		return lastRow, err
	}

	return lastRow, nil
}

// GetAllRows returns all the rows in the check table
func (s *Sqlite) GetAllRows() ([]gogios.Check, error) {
	data := []gogios.Check{}
	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Find(&data)

	return data, nil
}

// Init creates the database file and tables
func (s *Sqlite) Init() {
	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		helpers.Log.Println("Failed to connect to Sqlite to create database")
		helpers.Log.Printf("Error was: %s", err.Error())
	}
	defer db.Close()

	if !db.HasTable(&gogios.User{}) {
		db.AutoMigrate(&gogios.User{}, &gogios.Check{})
	}
}

func init() {
	databases.Add("sqlite", func() gogios.Database {
		return &Sqlite{}
	})
}
