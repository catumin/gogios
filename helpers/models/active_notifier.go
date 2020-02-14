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
