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

	"github.com/bkasin/gogios"
	_ "github.com/bkasin/gogios/databases/all"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/config"
	_ "github.com/bkasin/gogios/notifiers/all"
	"github.com/bkasin/gogios/web"
)

var (
	configFile = flag.String("config", "/etc/gogios/gogios.toml", "Config file to use")
	sampleConf = flag.Bool("sample_conf", false, "Print a sample config file to stdout")
)

func main() {
	flag.Parse()

	if *sampleConf {
		config.PrintSampleConfig()
		os.Exit(0)
	}

	// Create and start the log file
	helpers.Log.Printf("Gogios pid=%d", os.Getpid())

	// Read and print the config file
	conf := config.NewConfig()
	err := conf.GetConfig(*configFile)
	if err != nil {
		helpers.Log.Println("Check file could not be read, error return:")
		helpers.Log.Println(err.Error())
	}

	fmt.Println(conf.DatabaseNames())
	fmt.Println(conf.NotifierNames())

	// Need at least one database to start
	if len(conf.DatabaseNames()) == 0 {
		fmt.Println("gogios needs at least one database enabled to start.\nSqlite is the easiest to get started with.")
		os.Exit(1)
	}

	err = initPlugins(conf)
	if err != nil {
		fmt.Println("Could not initialize plugins. Error:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Start serving the website
	web.ServePage(conf)

	// Expose the REST API
	if conf.WebOptions.ExposeAPI {
		go web.API(conf)
	}

	// Set the PATH that will be used by checks
	os.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin:/usr/lib/gogios/plugins")

	// Do a round of checks immediately...
	runChecks(time.Now(), conf)
	// ... and then every $interval
	doEvery(conf.Options.Interval.Duration, runChecks, conf)
}

func runChecks(t time.Time, conf *config.Config) {
	// Read the raw check list into memory
	raw, err := ioutil.ReadFile("/etc/gogios/checks.json")
	if err != nil {
		helpers.Log.Println("Check file could not be read, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}

	// Create variables to hold the data for the currnet check list
	var curr []gogios.Check
	err = json.Unmarshal(raw, &curr)
	if err != nil {
		helpers.Log.Println("JSON could not be unmarshaled, error return:")
		helpers.Log.Println(err.Error())
		os.Exit(1)
	}

	// Use the first configured database as the primary for holding data
	primaryDB := conf.Databases[0].Database
	allPrev, err := primaryDB.GetAllRows()
	if err != nil {
		helpers.Log.Println("Could not read database")
		helpers.Log.Println(err.Error())
	}

	// Iterate through all the checks in the check list
	for i := 0; i < len(curr); i++ {
		curr[i].Status = "Failed"

		outputChannel := make(chan string, 1)
		go func() {
			commandReturn := check(curr[i])
			outputChannel <- commandReturn
		}()

		var goodCount = 0
		// Start at 1 because newly added checks will start as 1/0 or 0/0 otherwise
		var totalCount = 1

		prev, err := primaryDB.GetRow(curr[i], "title")
		if err != nil {
			helpers.Log.Println("Could not read database into prev variable")
			helpers.Log.Println(err.Error())
		}

		if prev.Title != "" {
			goodCount = prev.GoodCount
			totalCount = prev.TotalCount + 1
		}

		select {
		case output := <-outputChannel:
			if strings.Contains(output, curr[i].Expected) {
				curr[i].Status = "Success"
				goodCount++
			}
			curr[i].Output = output
		case <-time.After(conf.Options.Timeout.Duration):
			curr[i].Status = "Timed Out"
		}

		curr[i].Asof = time.Now()
		curr[i].GoodCount = goodCount
		curr[i].TotalCount = totalCount

		// Send out notifications through all enabled notifiers
		if prev.Title != "" && curr[i].Status != prev.Status {
			for _, notifier := range conf.Notifiers {
				err := notifier.Notifier.Notify(curr[i].Title, curr[i].Asof.Format(time.RFC822), curr[i].Output, curr[i].Status)
				if err != nil {
					helpers.Log.Println(err.Error())
				}
			}
		}

		// Set the current ID equal to the old ID, so that GORM can update the data properly
		// GORM will assign a new ID if prev.ID is nil
		curr[i].ID = prev.ID

		// Update or add rows for each configured database, then remove from allPrev[]
		for _, database := range conf.Databases {
			err := database.Database.AddRow(curr[i])
			if err != nil {
				helpers.Log.Println(err.Error())
			}

			for old := 0; old < len(allPrev); old++ {
				if allPrev[old].ID == curr[i].ID {
					allPrev = append(allPrev[:old], allPrev[old+1:]...)
				}
			}
		}

		if conf.Options.Verbose {
			err = helpers.AppendStringToFile("/var/log/gogios/service_check.log", curr[i].Asof.Format(time.RFC822)+" | Check "+curr[i].Title+" status: "+curr[i].Status)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
			err = helpers.AppendStringToFile("/var/log/gogios/service_check.log", "Output: \n"+curr[i].Output)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
		} else {
			err = helpers.AppendStringToFile("/var/log/gogios/service_check.log", curr[i].Asof.Format(time.RFC822)+" | Check "+curr[i].Title+" status: "+curr[i].Status)
			if err != nil {
				fmt.Println("Log could not be written. Error return:")
				fmt.Println(err.Error())
			}
		}
	}

	// Delete whatever is left in allPrev from the database
	for i := 0; i < len(allPrev); i++ {
		for _, database := range conf.Databases {
			err := database.Database.DeleteRow(allPrev[i], "id")
			if err != nil {
				helpers.Log.Println(err.Error())
			}
		}
	}
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
func doEvery(d time.Duration, f func(time.Time, *config.Config), conf *config.Config) {
	for x := range time.Tick(d) {
		f(x, conf)
	}
}

func check(check gogios.Check) string {
	var args = []string{"-c", check.Command}
	var output = getCommandOutput("/bin/sh", args)

	return output
}

// initPlugins calls the Init() function on any enabled notifiers and databases
func initPlugins(conf *config.Config) error {
	for _, d := range conf.Databases {
		err := d.Init()
		if err != nil {
			return fmt.Errorf("could not initialize database %s: %v", d.Config.Name, err)
		}
	}
	for _, n := range conf.Notifiers {
		err := n.Init()
		if err != nil {
			return fmt.Errorf("could not initialize notifier %s: %v", n.Config.Name, err)
		}
	}

	return nil
}
