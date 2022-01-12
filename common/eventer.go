package common

type Event interface {
	Stop()
}

type Eventer interface {
	Trigger(message string)
	Stop()
}
