package mysql

import (
	"errors"
	"fmt"
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

func (m *MySQL) openConnection() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", m.User+":"+m.Password+"@("+m.Host+":"+strconv.Itoa(m.Port)+")/"+m.Database+"?charset=utf8&parseTime=True&loc=Local")

	return db, err
}

// AddCheck makes sure an entry exists for the check and then adds to its history
func (m *MySQL) AddCheck(check gogios.Check, output string) error {
	db, err := m.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	data := gogios.CheckHistory{CheckID: &check.ID, Asof: &check.Asof, Output: output, Status: &check.Status}

	if db.NewRecord(check) {
		db.Create(&check)
		id, err := m.GetCheck(check.Title, "title")
		if err != nil {
			return err
		}
		data.CheckID = &id.ID
		db.Create(&data)
	} else {
		db.Model(check).Updates(&check)
		db.Create(&data)
	}

	return nil
}

// DeleteCheck will remove a row from the check table based on the ID
func (m *MySQL) DeleteCheck(check gogios.Check, field string) error {
	db, err := m.openConnection()
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

// GetCheck returns a single row. Searches using field (title or id) and returns
// the last record that matches
func (m *MySQL) GetCheck(searchField, searchType string) (gogios.Check, error) {
	lastRow := gogios.Check{}

	db, err := m.openConnection()
	if err != nil {
		return lastRow, err
	}
	defer db.Close()

	switch search := searchType; search {
	case "title":
		db.Where("title = ?", searchField).Last(&lastRow)
	case "id":
		db.Where("id = ?", searchField).Last(&lastRow)
	default:
		err = errors.New("searchType needs to be title or id")
		return lastRow, err
	}

	return lastRow, nil
}

// GetAllChecks returns all the rows in the check table
func (m *MySQL) GetAllChecks() ([]gogios.Check, error) {
	data := []gogios.Check{}
	db, err := m.openConnection()
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Where("deleted_at IS NULL").Find(&data)

	return data, nil
}

// GetCheckHistory returns $amount of rows of history for a check
func (m *MySQL) GetCheckHistory(check gogios.Check, amount int) ([]gogios.CheckHistory, error) {
	data := []gogios.CheckHistory{}
	db, err := m.openConnection()
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Raw("SELECT * FROM check_histories WHERE check_id = ? ORDER BY asof DESC LIMIT ?", check.ID, amount).Scan(&data)

	return data, nil
}

// AddUser inserts a new user into the database
func (m *MySQL) AddUser(user gogios.User) error {
	db, err := m.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if user.Username == "" || user.Password == "" {
		return errors.New("username or password was empty")
	}

	err = db.Create(&user).Error
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser sets the DeletedAt value of the user and blanks their password
func (m *MySQL) DeleteUser(user gogios.User) error {
	db, err := m.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	db.Where("Username = ?", user.Username).Update("password", "")
	db.Where("Username = ?", user.Username).Delete(&user)

	return nil
}

// GetUser looks up a user by username and returns their struct
func (m *MySQL) GetUser(user string) (*gogios.User, error) {
	data := gogios.User{}
	db, err := m.openConnection()
	if err != nil {
		return &data, err
	}
	defer db.Close()

	if err := db.Where("Username = ?", user).First(&data).Error; err != nil {
		fmt.Println(&data)
		return &data, err
	}

	return &data, nil
}

// Init creates the database file and tables
func (m *MySQL) Init() error {
	db, err := m.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if !db.HasTable(&gogios.CheckHistory{}) {
		db.AutoMigrate(&gogios.User{}, &gogios.Check{})
		db.AutoMigrate(&gogios.CheckHistory{}).AddForeignKey("check_id", "checks(id)", "RESTRICT", "RESTRICT")
	}

	return nil
}

func init() {
	databases.Add("mysql", func() gogios.Database {
		return &MySQL{}
	})
}
