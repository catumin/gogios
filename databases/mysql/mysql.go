package mysql

import (
	"errors"
	"strconv"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/databases"
	"github.com/jinzhu/gorm"

	// MySQL bindings for GORM
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// MySQL requirements
type MySQL struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

var sampleConfig = `
  ## MySQL server IP or address
  host = "127.0.0.1"
  port = 3306

  ## Username and password to authentication with
  user = ""
  password = ""

  ## Name of the database that will be used
  database = "gogios"
`

// SampleConfig returns the default config for MySQL
func (m *MySQL) SampleConfig() string {
	return sampleConfig
}

// Description returns a brief explanation of the database
func (m *MySQL) Description() string {
	return "Output check data to a MySQL database"
}

// AddCheckRow determines whether a record for the check exists in the database
// and either inserts a new row or updates the existing one
func (m *MySQL) AddCheckRow(check gogios.Check) error {
	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")
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
func (m *MySQL) DeleteCheckRow(check gogios.Check, field string) error {
	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")
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
func (m *MySQL) GetCheckRow(check gogios.Check, field string) (gogios.Check, error) {
	lastRow := gogios.Check{}

	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")
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
func (m *MySQL) GetAllCheckRows() ([]gogios.Check, error) {
	data := []gogios.Check{}
	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Find(&data)

	return data, nil
}

// Init creates the database file and tables
func (m *MySQL) Init() error {
	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")
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
	databases.Add("mysql", func() gogios.Database {
		return &MySQL{}
	})
}
