package gogios

type Notifier interface {
	SampleConfig() string

	Description() string

	Notify(check, time, output, status string) error
}
