package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	interval = flag.Int("interval", 3, "Time between check rounds")
)

// Check - struct to format checks
type Check struct {
	Title    string `json:"title"`
	Command  string `json:"command"`
	Expected string `json:"expected"`
	Good     bool   `json:"good"`
}

func main() {
	flag.Parse()
	os.Setenv("PATH", "/bin:/usr/bin:/sbin:/usr/local/bin")

	doEvery(time.Duration(*interval)*time.Minute, check)
}

func check(t time.Time) {
	raw, err := ioutil.ReadFile("/etc/gingertechengine/checks.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c []Check
	json.Unmarshal(raw, &c)

	for i := 0; i < len(c); i++ {
		var args = []string{"/bin/bash", "-c", c[i].Command}
		var output = getCommandOutput("sudo", args)
		fmt.Println("Check " + c[i].Title + " return: " + output)
		AppendStringToFile("/var/log/gingertechnology/service_check.log", time.Now().String()+" | Check "+c[i].Title+" return: "+output)
		if strings.Contains(output, c[i].Expected) {
			c[i].Good = true
		} else if !strings.Contains(output, c[i].Expected) {
			c[i].Good = false
		}
	}

	currentScore, _ := json.Marshal(c)
	err = ioutil.WriteFile("/opt/site/wwwroot/js/current.json", currentScore, 0644)
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
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
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
