package gogios

import (
	"time"

	"github.com/jinzhu/gorm"
)

// User - store users and their passwords
type User struct {
	gorm.Model

	Name     string `gorm:"size:255"`                // The user's name (Bailey Kasin)
	Username string `gorm:"size:30;unique;not null"` // The user's username (BKasin)
	Password string `gorm:"not null"`                // The password. Will be encrypted upon creation
}

// Check - track checks. Inactive checks will be marked with a DeletedAt datetime
type Check struct {
	gorm.Model

	Title      string    `gorm:"size:255;unique;not null"` // The name of the check
	Command    string    // The command that will be run in sh
	Expected   string    // Output that should be included in a succesful run of that check
	Status     string    // The most recent status. Success, Failed, Timed Our
	GoodCount  int       `json:"good_count"`  // The total number of times that this check has succeeded
	TotalCount int       `json:"total_count"` // The total number of times that this check has run
	Asof       time.Time `json:"asof"`        // Datetime that the most recent check finished at
}

// CheckHistory - stores the historical returns of each check that runs
type CheckHistory struct {
	gorm.Model

	CheckID *uint      `gorm:"ForeignKey:ID"`      // Foreign key. The ID of the check
	Asof    *time.Time `json:"asof"`               // Datetime that the check finished at
	Output  string     `gorm:"type:varchar(1250)"` // The output of the command that was run
	Status  *string    // The exit status of that check. Success, Failed, Timed Out
}

// Database object declaration
type Database interface {
	SampleConfig() string
	SubConfig() string

	Description() string

	AddCheck(check Check, output string) error
	DeleteCheck(check Check, field string) error
	GetCheck(searchField, searchType string) (Check, error)
	GetAllChecks() ([]Check, error)
	GetCheckHistory(check Check, amount int) ([]CheckHistory, error)
	AddUser(user User) error
	DeleteUser(user User) error
	GetUser(user string) (*User, error)
	// Init performs one time setup of the database and returns an error if the
	// configuration is invalid.
	Init() error
}
