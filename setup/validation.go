package setup

import (
	"regexp"
	"strings"
)

// CheckValid that the data entered into the setup form is valid
func (s *Setup) CheckValid() bool {
	// There must be a better way to do this, but I do not currently know it
	s.Errors = make(map[string]string)

	numberPatt := regexp.MustCompile(`\d`)
	yesnoPatt := regexp.MustCompile(`(yes|no)`)
	databasePatt := regexp.MustCompile(`(sqlite|mysql)`)
	ipPatt := regexp.MustCompile(`(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)
	ipDomainPatt := regexp.MustCompile(`((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)|((?:[\w-]+\.)*([\w-]{1,63})(?:\.(?:\w{3}|\w{2}))(?:$|\/)))`)

	if strings.TrimSpace(s.AdminUsername) == "" {
		s.Errors["username"] = "A username is required"
	}
	if strings.TrimSpace(s.AdminPassword) == "" {
		s.Errors["password"] = "A password field is required"
	}

	if !numberPatt.Match([]byte(s.Interval)) {
		s.Errors["interval"] = "This field must be a number"
	}
	if !numberPatt.Match([]byte(s.CheckTimeout)) {
		s.Errors["timeout"] = "This field must be a number"
	}

	if !ipPatt.Match([]byte(s.WebConfig.ListenIP)) {
		s.Errors["webip"] = "A valid IP must be entered"
	}
	if !numberPatt.Match([]byte(s.WebConfig.ListenPort)) {
		s.Errors["webport"] = "This field must be a number"
	}

	if !yesnoPatt.Match([]byte(s.WebConfig.ExposeAPI)) {
		s.Errors["api"] = "It must be specified whether the API will be available"
	} else {
		if s.WebConfig.ExposeAPI == "yes" {
			if !ipPatt.Match([]byte(s.WebConfig.ApiIP)) {
				s.Errors["apiip"] = "A valid IP must be entered"
			}
			if !numberPatt.Match([]byte(s.WebConfig.ApiPort)) {
				s.Errors["apiport"] = "This field must be a number"
			}
		}
	}

	if strings.TrimSpace(s.WebConfig.Title) == "" {
		s.Errors["title"] = "A site title is required"
	}

	if !databasePatt.Match([]byte(s.DatabasePreference)) {
		s.Errors["database"] = "A database must be selected from the dropdown list"
	} else {
		if s.DatabasePreference == "sqlite" {
			if strings.TrimSpace(s.Sqlite.Path) == "" {
				s.Errors["dbpath"] = "A path to a file writeable by gogios is required. The file should not exist yet, but the folders should"
			}
		} else if s.DatabasePreference == "mysql" {
			if !ipDomainPatt.Match([]byte(s.Mysql.Host)) {
				s.Errors["myurl"] = "A valid hostname or IP must be entered"
			}
			if !numberPatt.Match([]byte(s.Mysql.Port)) {
				s.Errors["myport"] = "A valid number must be entered"
			}
			if strings.TrimSpace(s.Mysql.User) == "" {
				s.Errors["myuser"] = "This field must be filled"
			}
			if strings.TrimSpace(s.Mysql.Password) == "" {
				s.Errors["mypass"] = "This field must be filled"
			}
			if strings.TrimSpace(s.Mysql.Database) == "" {
				s.Errors["mydb"] = "This field must be filled and the database must already be created on the server"
			}
		}
	}

	return len(s.Errors) == 0
}
