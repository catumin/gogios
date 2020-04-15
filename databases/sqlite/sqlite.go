package sqlite

import (
	"errors"
	"fmt"

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

func (s *Sqlite) openConnection() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", s.DBFile)

	return db, err
}

// AddCheck makes sure an entry exists for the check and then adds to its history
func (s *Sqlite) AddCheck(check gogios.Check, output string) error {
	db, err := s.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	data := gogios.CheckHistory{CheckID: &check.ID, Asof: &check.Asof, Output: output, Status: &check.Status}

	if db.NewRecord(check) {
		db.Create(&check)
		id, err := s.GetCheck(check.Title, "title")
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
func (s *Sqlite) DeleteCheck(check gogios.Check, field string) error {
	db, err := s.openConnection()
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
func (s *Sqlite) GetCheck(searchField, searchType string) (gogios.Check, error) {
	lastRow := gogios.Check{}

	db, err := s.openConnection()
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
func (s *Sqlite) GetAllChecks() ([]gogios.Check, error) {
	data := []gogios.Check{}
	db, err := s.openConnection()
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Where("deleted_at IS NULL").Find(&data)

	return data, nil
}

// GetCheckHistory returns $amount of rows of history for a check
func (s *Sqlite) GetCheckHistory(check gogios.Check, amount int) ([]gogios.CheckHistory, error) {
	data := []gogios.CheckHistory{}
	db, err := s.openConnection()
	if err != nil {
		return data, err
	}
	defer db.Close()

	db.Raw("SELECT * FROM check_histories WHERE check_id = ? ORDER BY asof DESC LIMIT ?", check.ID, amount).Scan(&data)

	return data, nil
}

// AddUser inserts a new user into the database
func (s *Sqlite) AddUser(user gogios.User) error {
	db, err := s.openConnection()
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
func (s *Sqlite) DeleteUser(user gogios.User) error {
	db, err := s.openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	db.Where("Username = ?", user.Username).Update("password", "")
	db.Where("Username = ?", user.Username).Delete(&user)

	return nil
}

// GetUser looks up a user by username and returns their struct
func (s *Sqlite) GetUser(user string) (*gogios.User, error) {
	data := gogios.User{}
	db, err := s.openConnection()
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
func (s *Sqlite) Init() error {
	db, err := s.openConnection()
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
	databases.Add("sqlite", func() gogios.Database {
		return &Sqlite{}
	})
}
