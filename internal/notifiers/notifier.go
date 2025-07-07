package notifiers

type Notifier interface {
	Notify(to, message string) error
}
