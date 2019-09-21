package helpers

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config - struct to hold the values read from the config file
type Config struct {
	Options    options
	WebOptions webOptions
	Telegram   telegram
	Twilio     twilio
}

type options struct {
	Interval int
	Verbose  bool
}

type webOptions struct {
	HTTPPort  int
	HTTPSPort int
	TLSCert   string
	TLSKey    string
	SSL       bool
	Redirect  bool
}

type telegram struct {
	API  string
	Chat string
}

type twilio struct {
	SID          string
	Token        string
	TwilioNumber string
	SendTo       string
}

// Version to be used by the web page
var Version = "1.4-dev"

// GetConfig reads and returns the confirg file as a struct
func GetConfig() Config {
	// Read and print the config file
	var conf Config
	if _, err := toml.DecodeFile("/etc/gingertechengine/gogios.toml", &conf); err != nil {
		fmt.Println("Config file could not be decoded, error return:")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%#v\n", conf)

	return conf
}
