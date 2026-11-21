package main

import (
	"fmt"
	"testing"
)

func TestHub(t *testing.T) {
	h := newHub()
	go h.run()

	a := &client{send: make(chan []byte, 16)}
	b := &client{send: make(chan []byte, 16)}

	h.register <- a
	h.register <- b
	h.unregister <- b

	h.broadcast <- []byte("light on")

	fmt.Println("a got: ", string(<-a.send))
	fmt.Println("b got: ", string(<-b.send))
}
