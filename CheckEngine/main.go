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
		fmt.Println(err)
	}
	fmt.Printf("%#v\n", conf)

	os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/local/bin")

	doEvery(time.Duration(conf.Options.Interval)*time.Minute, check, conf)
}

func check(t time.Time, conf Config) {
	raw, err := ioutil.ReadFile("/etc/gingertechengine/checks.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Check
	json.Unmarshal(raw, &c)

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
					AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | "+resp.Status)
				}
			}
		}

		fmt.Println("Check " + c[i].Title + " return: \n" + output)

		if conf.Options.verbose {
			AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | Check "+c[i].Title+" status: "+status)
			AppendStringToFile("/var/log/gingertechnology/service_check.log", "Output: \n"+output)
		} else {
			AppendStringToFile("/var/log/gingertechnology/service_check.log", c[i].Asof+" | Check "+c[i].Title+" status: "+status)
		}
	}

	currentStatus, _ := json.Marshal(c)
	err = ioutil.WriteFile("/opt/gingertechengine/js/current.json", currentStatus, 0644)
	if err != nil {
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
