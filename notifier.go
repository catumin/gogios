package gogios

type Notifier interface {
	SampleConfig() string

	Description() string

	Notify(check, time, output string, status bool) error
}

type Initializer interface {
	// Init performs one time setup of the notifier and returns an error if the
	// configuration is invalid.
	Init() error
}
