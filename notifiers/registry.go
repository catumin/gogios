package notifiers

import "github.com/bkasin/gogios"

type Creator func() gogios.Notifier

var Notifiers = map[string]Creator{}

func Add(name string, creator Creator) {
	Notifiers[name] = creator
}
