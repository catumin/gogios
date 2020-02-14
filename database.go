package gogios

import (
	"time"

	"github.com/jinzhu/gorm"
)

// User struct declaration
type User struct {
	gorm.Model

	Name     string `gorm:"size:255"`
	Username string `gorm:"size:255;unique;not null"`
	Password string `gorm:"not null"`
}

// Check - struct to format checks
type Check struct {
	gorm.Model

	Title      string `gorm:"size:255;unique;not null"`
	Command    string
	Expected   string
	Status     string
	GoodCount  int       `json:"good_count"`
	TotalCount int       `json:"total_count"`
	Asof       time.Time `json:"asof"`
	Output     string    `gorm:"type:varchar(1250)"`
}

type Database interface {
	SampleConfig() string

	Description() string

	AddRow(check Check) error
	DeleteRow(check Check, field string) error
	GetRow(check Check, field string) (Check, error)
	GetAllRows() ([]Check, error)
	// Init performs one time setup of the database and returns an error if the
	// configuration is invalid.
	Init() error
}
