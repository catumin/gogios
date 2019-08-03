package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config - struct to hold the values read from the config file
type Config struct {
	Options  options
	Telegram telegram
}

type options struct {
	Interval int
	Verbose  bool
}

type telegram struct {
	API  string
	Chat string
}

// Check - struct to format checks
type Check struct {
	Title    string `json:"title"`
	Command  string `json:"command"`
	Expected string `json:"expected"`
	Good     bool   `json:"good"`
	Asof     string `json:"asof"`
	Output   string `json:"output"`
}

func main() {
	// Read and print the config file
	var conf Config
	if _, err := toml.DecodeFile("/etc/gingertechengine/gogios.toml", &conf); err != nil {
		fmt.Println("Config file could not be decoded, error return:")
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", conf)

	// Set the PATH that will be used by checks
	os.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin")

	// Do a round of a checks immediately...
	check(time.Now(), conf)
	// ... and then every *interval
	doEvery(time.Duration(conf.Options.Interval)*time.Minute, check, conf)
}

func check(t time.Time, conf Config) {
	// Read the raw check list into memory
	raw, err := ioutil.ReadFile("/etc/gingertechengine/checks.json")
	if err != nil {
		fmt.Println("Check file could not be read, error return:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Create variables to hold the data for the currnet and previous check lists
	var curr, prev []Check
	err = json.Unmarshal(raw, &curr)
	if err != nil {
		fmt.Println("JSON could not be unmarshaled, error return:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Copy the check values from the previous round of checks to a different file...
	err = Copy("/opt/gingertechengine/js/current.json", "/opt/gingertechengine/js/prev.json")
	if err != nil {
		fmt.Println("Could not create copy of current check states, error return:")
		fmt.Println(err.Error())
	}

	// ... And then use that to set the prev variable to the old results
	raw, err = ioutil.ReadFile("/opt/gingertechengine/js/prev.json")
	if err != nil {
		fmt.Println("Previous check file could not be read, error return:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(raw, &prev)
	if err != nil {
		fmt.Println("JSON could not be unmarshaled, error return:")
		fmt.Println(err.Error())
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

		if len(prev) > i && curr[i].Good != prev[i].Good {
			if conf.Telegram.API != "" {
				err = PostMessage(conf.Telegram.API, conf.Telegram.Chat, curr[i].Title, curr[i].Asof, output, curr[i].Good)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}

		bytes := []byte("")
		curr[i].Output = string(strconv.AppendQuoteToASCII(bytes, output))

		fmt.Println("Check " + curr[i].Title + " return: \n" + output)

		if conf.Options.Verbose {
			err = AppendStringToFile("/var/log/gingertechnology/service_check.log", curr[i].Asof+" | Check "+curr[i].Title+" status: "+status)
			if err != nil {
				fmt.Println("Log could not be written. God save you, error return:")
				fmt.Println(err.Error())
			}
			err = AppendStringToFile("/var/log/gingertechnology/service_check.log", "Output: \n"+output)
			if err != nil {
				fmt.Println("Log could not be written. God save you, error return:")
				fmt.Println(err.Error())
			}
		} else {
			err = AppendStringToFile("/var/log/gingertechnology/service_check.log", curr[i].Asof+" | Check "+curr[i].Title+" status: "+status)
			if err != nil {
				fmt.Println("Log could not be written. God save you, error return:")
				fmt.Println(err.Error())
			}
		}
	}

	currentStatus, _ := json.Marshal(curr)
	err = ioutil.WriteFile("/opt/gingertechengine/js/current.json", currentStatus, 0644)
	if err != nil {
		fmt.Println("Result check file could not be written, error return:")
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v", curr)
}

func getCommandOutput(command string, args []string) (output string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
		return
	}
	sha := string(out)

	return sha
}

// doEvery - Run function f every d length of time
func doEvery(d time.Duration, f func(time.Time, Config), conf Config) {
	for x := range time.Tick(d) {
		f(x, conf)
	}
}
