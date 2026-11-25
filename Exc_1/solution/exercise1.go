package main

import "fmt"

type MessageData struct {
	Greeting string
	Target   string
}

func (m MessageData) PrintHelloWorld() {
	fmt.Printf("%s, %s!\n", m.Greeting, m.Target)
}

func main() {
	msg := MessageData{
		Greeting: "Hello",
		Target:   "World",
	}

	msg.PrintHelloWorld()
}