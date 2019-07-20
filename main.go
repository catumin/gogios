package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	verbose  bool
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
}

func main() {
	var conf Config
	if _, err := toml.DecodeFile("/etc/gingertechengine/gogios.toml", &conf); err != nil {
		fmt.Println("Config file could not be decoded, error return:")
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", conf)

	os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/local/bin")

	doEvery(time.Duration(conf.Options.Interval)*time.Minute, check, conf)
}

func check(t time.Time, conf Config) {
	raw, err := ioutil.ReadFile("/etc/gingertechengine/checks.json")
	if err != nil {
		fmt.Println("Check file could not be read, error return:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Check
	err = json.Unmarshal(raw, &c)
	if err != nil {
		fmt.Println("JSON could not be unmarshaled, error return:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for i := 0; i < len(c); i++ {
		var args = []string{"-c", c[i].Command}
		var output = getCommandOutput("/bin/sh", args)
		var status = "Failed"
		c[i].Asof = time.Now().Format(time.RFC822)

		if strings.Contains(output, c[i].Expected) {
			c[i].Good = true
			status = "Success"
		} else if !strings.Contains(output, c[i].Expected) {
			c[i].Good = false

			if conf.Telegram.API != "" {
				urlString := "https://api.telegram.org/bot" + conf.Telegram.API + "/sendMessage?chat_id=" + conf.Telegram.Chat + "&text=" + c[i].Title + " Status is Failed as of: " + c[i].Asof
				resp, err := http.Get(urlString)
				if err != nil {
					err = AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | "+resp.Status)
					if err != nil {
						fmt.Println("Log could not be written. God save you, error return:")
						fmt.Println(err.Error())
					}
				}
			}
		}

		fmt.Println("Check " + c[i].Title + " return: \n" + output)

		if conf.Options.verbose {
			err = AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | Check "+c[i].Title+" status: "+status)
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
			err = AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | Check "+c[i].Title+" status: "+status)
			if err != nil {
				fmt.Println("Log could not be written. God save you, error return:")
				fmt.Println(err.Error())
			}
		}
	}

	currentStatus, _ := json.Marshal(c)
	err = ioutil.WriteFile("/opt/gingertechengine/js/current.json", currentStatus, 0644)
	if err != nil {
		fmt.Println("Result check file could not be written, error return:")
		fmt.Println(err.Error())
	}
	fmt.Printf("%+v", c)
}

func getCommandOutput(command string, args []string) (output string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
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

// AppendStringToFile - Add string to the bottom of a file
func AppendStringToFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n" + text)
	if err != nil {
		return err
	}
	return nil
}
