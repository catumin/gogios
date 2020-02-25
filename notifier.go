package gogios

type Notifier interface {
	SampleConfig() string

	Description() string

	Notify(check, time, output, status string) error

	// Init performs one time setup of the notifier and returns an error if the
	// configuration is invalid.
	Init() error
}
