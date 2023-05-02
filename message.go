package main

type Message struct {
}

func (m *Message) encode() []byte {
	return make([]byte, 256)
}
