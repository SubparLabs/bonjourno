package inputs

import (
	"math/rand"
	"time"
)

// MessageChooser - Chooses one message from a list
type MessageChooser <-chan string

func SequentialMessageChooser(builder MessageBuilder) MessageChooser {
	c := make(chan string)

	go func() {
		var messages []string
		index := -1

		for {
			select {
			case messages = <-builder:
			default:
			}

			if len(messages) > 0 {
				index = (index + 1) % len(messages)
				c <- messages[index]
			}

		}
	}()

	return c
}

func RandomMessageChooser(builder MessageBuilder) MessageChooser {
	c := make(chan string)

	go func() {
		var messages []string

		for {
			select {
			case messages = <-builder:
			default:
			}

			if len(messages) > 0 {
				c <- messages[rand.Intn(len(messages))]
			}
		}
	}()

	return c
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
