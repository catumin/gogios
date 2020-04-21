package config

import "fmt"

// InitPlugins calls the Init() function on any enabled notifiers and databases
func InitPlugins() error {
	for _, d := range Conf.Databases {
		err := d.Database.Init()
		if err != nil {
			return fmt.Errorf("could not initialize database %s: %v", d.Config.Name, err)
		}
	}
	for _, n := range Conf.Notifiers {
		err := n.Notifier.Init()
		if err != nil {
			return fmt.Errorf("could not initialize notifier %s: %v", n.Config.Name, err)
		}
	}

	return nil
}
