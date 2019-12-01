package helpers

import (
	"fmt"
	"strconv"

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
	IP        string
	HTTPPort  int
	HTTPSPort int
	TLSCert   string
	TLSKey    string
	SSL       bool
	Redirect  bool
	ExposeAPI bool
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
var Version = "1.4"

// ConfigTest reads and tests a config file to make sure
// that all options are valid
func ConfigTest(conf Config) {
	fmt.Printf("OPTIONS\nGogios will run checks every: %d minutes\nVerbose logging is set to: %s\n", conf.Options.Interval, strconv.FormatBool(conf.Options.Verbose))
	fmt.Printf("WEB OPTIONS\nGogios will listen on IP: %s\nHTTP port is set to: %d\n", conf.WebOptions.IP, conf.WebOptions.HTTPPort)

	if conf.WebOptions.SSL {
		fmt.Printf("Gogios will listen for HTTPS on port: %d\nUsing TLS Cert: %s\nAnd TLS Key: %s\n", conf.WebOptions.HTTPSPort, conf.WebOptions.TLSCert, conf.WebOptions.TLSKey)
		if conf.WebOptions.Redirect {
			fmt.Println("Gogios will attempt to redirect HTTP to HTTPS")
		} else {
			fmt.Println("Gogios will not attempt to redirect HTTP to HTTPS")
		}
	} else {
		fmt.Println("Gogios will not listen on HTTPS")
	}

	if conf.WebOptions.ExposeAPI {
		fmt.Println("The Gogios web API will be available")
	}

	fmt.Printf("NOTIFIERS\nTELEGRAM\n")

	if conf.Telegram.API != "" {
		fmt.Printf("Gogios will attempt to send Telegram messages to Chat: %s\nUsing Bot ID: %s\n", conf.Telegram.Chat, conf.Telegram.API)
	} else {
		fmt.Println("Gogios will not attempt to send Telegram messages")
	}

	fmt.Println("TWILIO")

	if conf.Twilio.SID != "" {
		fmt.Printf("Gogios will attempt to use Twilio to send messages from: %s\nTo: %s\nWith account SID: %s\nAnd auth token: %s\n", conf.Twilio.TwilioNumber, conf.Twilio.SendTo, conf.Twilio.SID, conf.Twilio.Token)
	} else {
		fmt.Println("Gogios will not attempt to send Twilio messages")
	}
}

// GetConfig reads and returns the confirg file as a struct
func GetConfig(config string) (Config, error) {
	// Read and print the config file
	var conf Config

	if _, err := toml.DecodeFile(config, &conf); err != nil {
		fmt.Printf("Config file could not be decoded, error return:\n%s", err.Error())
		return conf, err
	}

	ConfigTest(conf)

	return conf, nil
}
