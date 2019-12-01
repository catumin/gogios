package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/notifiers"
	"github.com/bkasin/gogios/web"
)

var (
	config = flag.String("config", "/etc/gingertechengine/gogios.toml", "Config file to use")
)

// Check - struct to format checks
type Check struct {
	ID       int    `json:id`
	Title    string `json:"title"`
	Command  string `json:"command"`
	Expected string `json:"expected"`
	Good     bool   `json:"good"`
	Asof     string `json:"asof"`
	Output   string `json:"output"`
}

func main() {
	flag.Parse()

	// Create and start the log file
	helpers.Log.Printf("Gogios pid=%d", os.Getpid())

	// Read and print the config file
	conf, err := helpers.GetConfig(*config)
	if err != nil {
		helpers.Log.Printf("Gogios config test failed. Error was:\n%s", err.Error())
		os.Exit(1)
	}

	// Start serving the website
	web.ServePage(conf)

	// Set the PATH that will be used by checks
	os.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin")

	// Do a round of checks immediately...
	check(time.Now(), conf)
	// ... and then every $interval
	doEvery(time.Duration(conf.Options.Interval)*time.Minute, check, conf)
}

func check(t time.Time, conf helpers.Config) {
	// Read the raw check list into memory
	raw, err := ioutil.ReadFile("/etc/gingertechengine/checks.json")
	if err != nil {
		helpers.Log.Println("Check file could not be read, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}

	// Create variables to hold the data for the currnet and previous check lists
	var curr, prev []Check
	err = json.Unmarshal(raw, &curr)
	if err != nil {
		helpers.Log.Println("JSON could not be unmarshaled, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}

	// Copy checks.json to current.json if it does not exist
	if _, err := os.Stat("/opt/gingertechengine/js/current.json"); os.IsNotExist(err) {
		err = helpers.Copy("/etc/gingertechengine/checks.json", "/opt/gingertechengine/js/current.json")
		if err != nil {
			helpers.Log.Println("Could not copy checks template to current.json, error return:")
			helpers.Log.Println(err.Error())
		}
	}

	// Copy the check values from the previous round of checks to a different file...
	err = helpers.Copy("/opt/gingertechengine/js/current.json", "/opt/gingertechengine/js/prev.json")
	if err != nil {
		helpers.Log.Println("Could not create copy of current check states, error return:")
		helpers.Log.Println(err.Error())
	}

	// ... And then use that to set the prev variable to the old results
	raw, err = ioutil.ReadFile("/opt/gingertechengine/js/prev.json")
	if err != nil {
		helpers.Log.Println("Previous check file could not be read, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(raw, &prev)
	if err != nil {
		helpers.Log.Println("JSON could not be unmarshaled, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}

	// Iterate through all the checks in the check list
	for i := 0; i < len(curr); i++ {
		var args = []string{"-c", curr[i].Command}
		var output = getCommandOutput("/bin/sh", args)
		var status = "Failed"
		curr[i].Asof = time.Now().Format(time.RFC822)

		if strings.Contains(output, curr[i].Expected) {
			curr[i].Good = true
			status = "Success"
		} else if !strings.Contains(output, curr[i].Expected) {
			curr[i].Good = false
		}

		// Send out notifications if requested
		if len(prev) > i && curr[i].Good != prev[i].Good {
			if conf.Telegram.API != "" {
				err = notifiers.TelegramMessage(conf.Telegram.API, conf.Telegram.Chat, curr[i].Title, curr[i].Asof, output, curr[i].Good)
				if err != nil {
					helpers.Log.Println(err.Error())
				}
			}

			if conf.Twilio.Token != "" {
				err = notifiers.TwilioMessage(conf.Twilio.SID, conf.Twilio.Token, conf.Twilio.TwilioNumber, conf.Twilio.SendTo, curr[i].Title, curr[i].Asof, output, curr[i].Good)
				if err != nil {
					helpers.Log.Println(err.Error())
				}
			}
		}

		err = helpers.WriteStringToFile("/opt/gingertechengine/js/output/"+curr[i].Title, output)
		if err != nil {
			helpers.Log.Printf("Output for check %s could not be written to output file. Error return: %s", curr[i].Title, err.Error())
		}

		helpers.Log.Println("Check " + curr[i].Title + " return: \n" + output)

		if conf.Options.Verbose {
			err = helpers.AppendStringToFile("/var/log/gingertechnology/service_check.log", curr[i].Asof+" | Check "+curr[i].Title+" status: "+status)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
			err = helpers.AppendStringToFile("/var/log/gingertechnology/service_check.log", "Output: \n"+output)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
		} else {
			err = helpers.AppendStringToFile("/var/log/gingertechnology/service_check.log", curr[i].Asof+" | Check "+curr[i].Title+" status: "+status)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
		}
	}

	currentStatus, _ := json.Marshal(curr)
	err = ioutil.WriteFile("/opt/gingertechengine/js/current.json", currentStatus, 0644)
	if err != nil {
		helpers.Log.Println("Result check file could not be written, error return:")
		helpers.Log.Println(err.Error())
	}
	helpers.Log.Printf("%+v", curr)
}

func getCommandOutput(command string, args []string) (output string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		helpers.Log.Printf("cmd.Run() failed with %s\n", err)
		return
	}
	sha := string(out)

	return sha
}

// doEvery - Run function f every d length of time
func doEvery(d time.Duration, f func(time.Time, helpers.Config), conf helpers.Config) {
	for x := range time.Tick(d) {
		f(x, conf)
	}
}
