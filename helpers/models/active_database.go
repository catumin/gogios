package models

import "github.com/bkasin/gogios"

type ActiveDatabase struct {
	Database gogios.Database
	Config   *DatabaseConfig
}

type DatabaseConfig struct {
	Name  string
	Alias string
}

func NewActiveDatabase(database gogios.Database, config *DatabaseConfig) *ActiveDatabase {
	return &ActiveDatabase{
		Database: database,
		Config:   config,
	}
}

func (d *ActiveDatabase) Init() error {
	if p, ok := d.Database.(gogios.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}

	return nil
}
