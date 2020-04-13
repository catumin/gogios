package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bkasin/gogios/databases"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/models"
	"github.com/bkasin/gogios/notifiers"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

// Config - struct to hold the values read from the config file
type Config struct {
	Options    *OptionsConfig
	WebOptions *WebOptionsConfig

	Notifiers []*models.ActiveNotifier
	Databases []*models.ActiveDatabase
}

// OptionsConfig - General system options such as check interval
type OptionsConfig struct {
	// Interval on which to check in minutes
	Interval helpers.Duration
	// Verbose controls whether check output will be logged, and how much will be logged to stdout
	Verbose bool

	// Timeout for each check
	Timeout helpers.Duration
}

// WebOptionsConfig - Options related to the web interface
type WebOptionsConfig struct {
	// IP to listen on
	IP string
	// Port to use for non-SSL connections
	HTTPPort int `toml:"http_port"`
	// Port to use for SSL connections
	HTTPSPort int `toml:"https_port"`

	// TLS settings. Cert, key, and whether to listen on SSL
	TLSCert string `toml:"tls_cert"`
	TLSKey  string `toml:"tls_key"`
	SSL     bool
	// Redirect to SSL
	Redirect bool
	// Allow the REST API to be accessible
	ExposeAPI bool `toml:"expose_api"`
	// IP to listen for API connections on
	APIIP string `toml:"api_ip"`
	// Port to listen for API connections on
	APIPort int `toml:"api_port"`

	// Web interface branding
	Title  string
	NavBar string `toml:"nav_bar"`
	Logo   string
}

// NewConfig provides base config options that get replaced by the TOML options
func NewConfig() *Config {
	c := &Config{
		Options: &OptionsConfig{
			Interval: helpers.Duration{Duration: 3 * time.Minute},
			Verbose:  false,
			Timeout:  helpers.Duration{Duration: 60 * time.Second},
		},

		WebOptions: &WebOptionsConfig{
			IP:        "127.0.0.1",
			HTTPPort:  8411,
			HTTPSPort: 8412,
			TLSCert:   "",
			TLSKey:    "",
			SSL:       false,
			Redirect:  false,
			ExposeAPI: true,
			APIIP:     "127.0.0.1",
			APIPort:   8413,
			Title:     "Ginger Technology Service Check Engine",
			NavBar:    "Ginger Technology Service Check Engine",
			Logo:      "/usr/share/gogios/logo.jpg",
		},
		Notifiers: make([]*models.ActiveNotifier, 0),
		Databases: make([]*models.ActiveDatabase, 0),
	}

	return c
}

// GetConfig reads and parses the config
func (c *Config) GetConfig(config string) error {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return fmt.Errorf("Read error %s, %s", config, err)
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		return fmt.Errorf("Parse error %s, %s", config, err)
	}

	if val, ok := tbl.Fields["options"]; ok {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid config", config)
		}
		if err = toml.UnmarshalTable(subTable, c.Options); err != nil {
			log.Printf("Could not parse [options] config\n")
			return fmt.Errorf("Error parsing %s, %s", config, err)
		}
	}

	if val, ok := tbl.Fields["web_options"]; ok {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid config", config)
		}
		if err = toml.UnmarshalTable(subTable, c.WebOptions); err != nil {
			log.Printf("Could not parse [web_options] config\n")
			return fmt.Errorf("Error parsing %s, %s", config, err)
		}
	}

	// Notifiers and Databases
	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid config", config)
		}

		switch name {
		case "options", "web_options":
		case "notifiers":
			for notifierName, val := range subTable.Fields {
				switch notifierSubTable := val.(type) {
				case []*ast.Table:
					for _, t := range notifierSubTable {
						if err = c.addNotifier(notifierName, t); err != nil {
							return fmt.Errorf("Error parsing %s, %s", config, err)
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s, file %s", notifierName, config)
				}
			}
		case "databases":
			for databaseName, val := range subTable.Fields {
				switch databaseSubTable := val.(type) {
				case []*ast.Table:
					for _, t := range databaseSubTable {
						if err = c.addDatabase(databaseName, t); err != nil {
							return fmt.Errorf("Error parsing: %s, %s", config, err)
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s, file %s", databaseName, config)
				}
			}
		default:
			fmt.Printf("Unrecognized config option: %s", name)
		}
	}

	return nil
}

// NotifierNames returns a list of all configured notifiers
func (c *Config) NotifierNames() []string {
	var name []string
	for _, notifier := range c.Notifiers {
		name = append(name, notifier.Config.Name)
	}

	return name
}

// DatabaseNames returns a list of all configured databases
func (c *Config) DatabaseNames() []string {
	var name []string
	for _, database := range c.Databases {
		name = append(name, database.Config.Name)
	}

	return name
}

var header = `# Options for Gogios
# https://github.com/bkasin/gogios
# https://angrysysadmins.tech

`

var optionsConfig = `
[options]
  # How often to run checks in minutes
  interval = "3m"

  # Include check output in the log file, and increase
  # how much information is sent to standard out
  verbose = false

  # Per check timeout in seconds
  # If enough checks get stuck it is possible that the
  # next round will start before the previous finishes
  timeout = "60s"

`

var webConfig = `
[web_options]
  # Change IP to 0.0.0.0 to listen on all interfaces
  IP = "127.0.0.1"
  http_port = 8411
  https_port = 8412

  # Should the website be hosted on HTTPS
  SSL = false
  # Redirect from HTTP to HTTPS
  redirect = false

  # Path to TLS cert and key for HTTPS
  tls_cert = ""
  tls_key = ""

  # Allow the REST API to be accessed on the IP and port
  # specified
  expose_api = true
  api_ip = "127.0.0.1"
  api_port = 8413

  # BRANDING
  # The options in this section will alter the branding on the web interface

  # The text that shows after the current page in the web browser's task bar
  title = "Ginger Technology Service Check Engine"

  # The text that shows before the tabs on the navigation bar
  nav_bar = "Ginger Technology Service Check Engine"

  # A small logo that can appear in the navigation bar
  logo = "/usr/share/gogios/logo.jpg"

`

var databaseHeader = `
###########################
#
# Databases
#
###########################

`

var notifierHeader = `
###########################
#
# Notifiers
#
###########################

`

// PrintSampleConfig prints the sample config
func PrintSampleConfig() {
	fmt.Printf(header)
	fmt.Printf(optionsConfig)
	fmt.Printf(webConfig)

	fmt.Printf(databaseHeader)
	printDatabases(true)

	fmt.Printf(notifierHeader)
	printNotifiers(true)
}

func printNotifiers(commented bool) {
	var anames []string
	for aname := range notifiers.Notifiers {
		anames = append(anames, aname)
	}
	sort.Strings(anames)

	for _, aname := range anames {
		creator := notifiers.Notifiers[aname]
		notifier := creator()

		printConfig(aname, notifier, "notifiers", commented)
	}
}

func printDatabases(commented bool) {
	var anames []string
	for aname := range databases.Databases {
		anames = append(anames, aname)
	}
	sort.Strings(anames)

	for _, aname := range anames {
		creator := databases.Databases[aname]
		database := creator()

		printConfig(aname, database, "databases", commented)
	}
}

type data interface {
	Description() string
	SampleConfig() string
}

func printConfig(name string, d data, cat string, commented bool) {
	comment := ""
	if commented {
		comment = "# "
	}
	fmt.Printf("\n%s# %s\n%s[[%s.%s]]", comment, d.Description(), comment, cat, name)

	config := d.SampleConfig()
	if config == "" {
		fmt.Printf("\n%s  # no configuration\n\n", comment)
	} else {
		lines := strings.Split(config, "\n")
		for i, line := range lines {
			if i == 0 || i == len(lines)-1 {
				fmt.Print("\n")
				continue
			}
			fmt.Print(strings.TrimRight(comment+line, " ") + "\n")
		}
	}
}

func (c *Config) addNotifier(name string, table *ast.Table) error {
	creator, ok := notifiers.Notifiers[name]
	if !ok {
		return fmt.Errorf("Undefined but requested notifier: %s", name)
	}
	notifier := creator()

	notifierConfig, err := buildNotifier(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, notifier); err != nil {
		return err
	}

	rf := models.NewActiveNotifier(notifier, notifierConfig)

	c.Notifiers = append(c.Notifiers, rf)

	return nil
}

func (c *Config) addDatabase(name string, table *ast.Table) error {
	creator, ok := databases.Databases[name]
	if !ok {
		return fmt.Errorf("Undefined but requested database: %s", name)
	}
	database := creator()

	databaseConfig, err := buildDatabase(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, database); err != nil {
		return err
	}

	rf := models.NewActiveDatabase(database, databaseConfig)

	c.Databases = append(c.Databases, rf)

	return nil
}

func buildNotifier(name string, tbl *ast.Table) (*models.NotifierConfig, error) {
	conf := &models.NotifierConfig{Name: name}

	return conf, nil
}

func buildDatabase(name string, tbl *ast.Table) (*models.DatabaseConfig, error) {
	conf := &models.DatabaseConfig{Name: name}

	return conf, nil
}
