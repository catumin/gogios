package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/api"
	_ "github.com/bkasin/gogios/databases/all"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/config"
	_ "github.com/bkasin/gogios/notifiers/all"
	"github.com/bkasin/gogios/setup"
	"github.com/bkasin/gogios/web"
	"github.com/google/logger"
)

var (
	configFile = flag.String("config", "/etc/gogios/gogios.toml", "Config file to use")
	sampleConf = flag.Bool("sample_conf", false, "Print a sample config file to stdout")
	notify     = flag.String("notify", "", "Send a message to all notifiers")
)

func main() {
	flag.Parse()

	if *sampleConf {
		config.PrintSampleConfig()
		os.Exit(0)
	}

	// Read and print the config file
	config.Conf = config.NewConfig()

	// Prepare the logger for initialization tasks
	log, err := os.OpenFile("/var/log/gogios/init.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	initialLogger := logger.Init("InitialLog", config.Conf.Options.Verbose, true, log)
	defer initialLogger.Close()
	initialLogger.Infof("Gogios pid=%d", os.Getpid())

	err = config.Conf.GetConfig(*configFile)
	if err != nil {
		initialLogger.Errorf("Config file could not be read, error return:\n%s", err.Error())
	}

	initialLogger.Infoln(config.Conf.DatabaseNames())
	initialLogger.Infoln(config.Conf.NotifierNames())

	// If -notify is set, then send the message and exit before checking databases
	if *notify != "" {
		for _, notifier := range config.Conf.Notifiers {
			err := notifier.Notifier.Notify("External Message", time.Now().Format(time.RFC822), *notify, "Send")
			if err != nil {
				initialLogger.Errorln(err.Error())
				os.Exit(1)
			}
		}
		os.Exit(1)
	}

	// Need at least one database to start
	if len(config.Conf.DatabaseNames()) == 0 {
		initialLogger.Errorln("gogios needs at least one database enabled to start.\nSetup webpage is being exposed.")
		setup.FirstSetup(*configFile)

		fmt.Println("Passed setup")
	}

	err = config.InitPlugins()
	if err != nil {
		initialLogger.Errorf("Could not initialize plugins. Error:\n%s", err.Error())
		os.Exit(1)
	}

	// Set the PATH that will be used by checks
	os.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin:/usr/lib/gogios/plugins")

	// Do a round of checks immediately...
	runChecks(time.Now())

	// ... and then every $interval
	go doEvery(config.Conf.Options.Interval.Duration, runChecks)

	// Expose the REST API
	if config.Conf.WebOptions.ExposeAPI {
		go api.API()
	}

	// Start serving the website
	web.ServePage()
}

func runChecks(t time.Time) {
	// Start the service check logger
	log, err := os.OpenFile("/var/log/gogios/service_check.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer log.Close()

	checkLogger := logger.Init("ServiceCheck", config.Conf.Options.Verbose, true, log)
	defer checkLogger.Close()

	// Read the raw check list into memory
	raw, err := os.ReadFile("/etc/gogios/checks.json")
	if err != nil {
		checkLogger.Errorf("Check file could not be read, error return:\n%s", err.Error())
		os.Exit(1)
	}

	// Create variables to hold the data for the currnet check list
	var curr []gogios.Check
	err = json.Unmarshal(raw, &curr)
	if err != nil {
		checkLogger.Errorf("JSON could not be unmarshaled, error return:\n%s", err.Error())
		os.Exit(1)
	}

	// Use the first configured database as the primary for holding data
	primaryDB := config.Conf.Databases[0].Database
	allPrev, err := primaryDB.GetAllChecks()
	if err != nil {
		checkLogger.Errorf("Could not read database, error return:\n%s", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(len(curr))

	// Iterate through all the checks in the check list
	for i := 0; i < len(curr); i++ {
		go func(i int) {
			defer wg.Done()
			curr[i].Status = "Failed"

			outputChannel := make(chan string, 1)
			go func() {
				commandReturn := check(checkLogger, curr[i])
				outputChannel <- commandReturn
			}()

			var goodCount = 0
			// Start at 1 because newly added checks will start as 1/0 or 0/0 otherwise
			var totalCount = 1

			prev, err := primaryDB.GetCheck(curr[i].Title, "title")
			if err != nil {
				checkLogger.Errorf("Could not read database into prev variable, error return:\n%s", err.Error())
			}

			if prev.Title != "" {
				goodCount = prev.GoodCount
				totalCount = prev.TotalCount + 1
			}

			var Output string
			select {
			case output := <-outputChannel:
				if strings.Contains(output, curr[i].Expected) {
					curr[i].Status = "Success"
					goodCount++
				}
				Output = output
			case <-time.After(config.Conf.Options.Timeout.Duration):
				curr[i].Status = "Timed Out"
			}

			curr[i].Asof = time.Now()
			curr[i].GoodCount = goodCount
			curr[i].TotalCount = totalCount

			// Send out notifications through all enabled notifiers
			if prev.Title != "" && curr[i].Status != prev.Status {
				for _, notifier := range config.Conf.Notifiers {
					err := notifier.Notifier.Notify(curr[i].Title, curr[i].Asof.Format(time.RFC822), Output, curr[i].Status)
					if err != nil {
						checkLogger.Errorln(err.Error())
					}
				}
			}

			// Set the current ID equal to the old ID, so that GORM can update the data properly
			// GORM will assign a new ID if prev.ID is nil
			curr[i].ID = prev.ID

			// Update or add rows for each configured database, then remove from allPrev[]
			for _, database := range config.Conf.Databases {
				err := database.Database.AddCheck(curr[i], Output)
				if err != nil {
					checkLogger.Errorln(err.Error())
				}

				for old := 0; old < len(allPrev); old++ {
					if allPrev[old].ID == curr[i].ID {
						allPrev = append(allPrev[:old], allPrev[old+1:]...)
					}
				}
			}

			checkLogger.Infof("Check %s status: %s as of: %s\n", curr[i].Title, curr[i].Status, curr[i].Asof.Format(time.RFC822))
			if config.Conf.Options.Verbose {
				checkLogger.Infof("Output: \n%s", Output)
			}

			web.UpdateWebData()
		}(i)
	}

	wg.Wait()

	// Delete whatever is left in allPrev from the database
	for i := 0; i < len(allPrev); i++ {
		for _, database := range config.Conf.Databases {
			err := database.Database.DeleteCheck(allPrev[i], "id")
			if err != nil {
				checkLogger.Errorln(err.Error())
			}
		}
	}
}

// doEvery - Run function f every d length of time
func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func check(logger *logger.Logger, check gogios.Check) string {
	var args = []string{"-c", check.Command}
	var output = helpers.GetCommandOutput(logger, "/bin/sh", args)

	return output
}
