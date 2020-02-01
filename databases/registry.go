package databases

import "github.com/bkasin/gogios"

type Creator func() gogios.Database

var Databases = map[string]Creator{}

func Add(name string, creator Creator) {
	Databases[name] = creator
}
