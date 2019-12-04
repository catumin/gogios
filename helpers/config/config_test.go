package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	conf, err := GetConfig("../package_files/gogios.sample.toml")

	if err != nil {
		t.Errorf("Config test failed, got error: %s", err)
	}

	if conf.Options.Interval != 3 {
		t.Errorf("Config contained bad value. Expected 3, got: %d", conf.Options.Interval)
	}
}
