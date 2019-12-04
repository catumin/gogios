package models

import "github.com/bkasin/gogios"

type ActiveNotifier struct {
	Notifier gogios.Notifier
	Config   *NotifierConfig
}

type NotifierConfig struct {
	Name  string
	Alias string
}

func NewActiveNotifier(notifier gogios.Notifier, config *NotifierConfig) *ActiveNotifier {
	return &ActiveNotifier{
		Notifier: notifier,
		Config:   config,
	}
}

func (a *ActiveNotifier) Init() error {
	if p, ok := a.Notifier.(gogios.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}

	return nil
}
