package setup

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bkasin/gogios"
	"github.com/bkasin/gogios/helpers"
	"github.com/bkasin/gogios/helpers/config"
	"github.com/bkasin/gogios/users"
)

func confirmConfig(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Confirm config")
	render(w, "confirm_config.html", printReadyConfig, setupLogger)
}

func finalizeSetup(setupConfig string, setup Setup) error {
	err := helpers.WriteStringToFile(configPath, setupConfig)
	if err != nil {
		return err
	}
	err = config.Conf.GetConfig(configPath)
	if err != nil {
		return err
	}
	err = config.InitPlugins()
	if err != nil {
		setupLogger.Errorf("Could not initialize plugins. Error:\n%s", err.Error())
		os.Exit(1)
	}

	admin := gogios.User{
		Name:     setup.AdminName,
		Username: setup.AdminUsername,
		Password: setup.AdminPassword,
	}

	err = users.CreateUser(admin, config.Conf)
	if err != nil {
		return err
	}

	return nil
}
