package utils

type Message struct {
	Body []byte
	Ack  func(multiple bool) error
	Nack func(multiple bool, requeue bool) error
}
