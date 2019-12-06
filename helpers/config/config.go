package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"

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
}

// OptionsConfig - General system options such as check interval
type OptionsConfig struct {
	// Interval on which to check in minutes
	Interval int
	// Verbose controls whether check output will be logged
	Verbose bool
}

// WebOptionsConfig - Options related to the web interface
type WebOptionsConfig struct {
	// IP to listen on
	IP string
	// Port to use for non-SSL connections
	HTTPPort int
	// Port to use for SSL connections
	HTTPSPort int

	// TLS settings. Cert, key, and whether to listen on SSL
	TLSCert string
	TLSKey  string
	SSL     bool
	// Redirect to SSL
	Redirect bool
	// Allow the REST API to be accessible
	ExposeAPI bool
	// IP to listen for API connections on
	APIIP string
	// Port to listen for API connections on
	APIPort int
}

// NewConfig provides base config options that get replaced by the TOML options
func NewConfig() *Config {
	c := &Config{
		Options: &OptionsConfig{
			Interval: 3,
			Verbose:  false,
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
			APIIP:     "0.0.0.0",
			APIPort:   8413,
		},
		Notifiers: make([]*models.ActiveNotifier, 0),
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
			return fmt.Errorf("%s: invalid configuration", config)
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

	// Notifiers
	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid config", config)
		}

		switch name {
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
		default:
			fmt.Println("Unrecognized config option: %", name)
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

var header = `# Options for Gogios
# https://github.com/bkasin/gogios
# https://angrysysadmins.tech

`

var optionsConfig = `
[options]
  # How often to run checks in minutes
  interval = 3
  # Verbose logging. true or false
  verbose = false

`

var webConfig = `
[web_options]
  # Change IP to 0.0.0.0 to listen on all interfaces
  IP = "127.0.0.1"
  HTTPPort = 8411
  HTTPSPort = 8412

  # Should the website be hosted on HTTPS
  SSL = false
  # Redirect from HTTP to HTTPS
  redirect = false

  # Path to TLS cert and key for HTTPS
  TLSCert = ""
  TLSKey = ""

  exposeAPI = true
  APIIP = "0.0.0.0"
  APIPort = 8413

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

func buildNotifier(name string, tbl *ast.Table) (*models.NotifierConfig, error) {
	conf := &models.NotifierConfig{Name: name}

	return conf, nil
}
