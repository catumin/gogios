package sqlite

import (
	"errors"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/databases"
	"github.com/jinzhu/gorm"

	// SQLite bindings for GORM
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

// AddCheckRow determines whether a record for the check exists in the database
// and either inserts a new row or updates the existing one
func (s *Sqlite) AddCheckRow(check gogios.Check) error {
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

// DeleteCheckRow will remove a row from the check table based on the ID
func (s *Sqlite) DeleteCheckRow(check gogios.Check, field string) error {
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

// GetCheckRow returns a single row. Searches using field (title or id) and returns
// the last record that matches
func (s *Sqlite) GetCheckRow(check gogios.Check, field string) (gogios.Check, error) {
	lastRow := gogios.Check{}

	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return lastRow, err
	}
	defer db.Close()

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

// GetAllCheckRows returns all the rows in the check table
func (s *Sqlite) GetAllCheckRows() ([]gogios.Check, error) {
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
func (s *Sqlite) Init() error {
	db, err := gorm.Open("sqlite3", s.DBFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if !db.HasTable(&gogios.User{}) {
		db.AutoMigrate(&gogios.User{}, &gogios.Check{})
	}

	return nil
}

func init() {
	databases.Add("sqlite", func() gogios.Database {
		return &Sqlite{}
	})
}
