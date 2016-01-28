package inputs

import (
	"bufio"
	"fmt"
	"strings"
)

// MessageBuilder - Turns raw data into a list of messages
type MessageBuilder <-chan []string

func Lines(source DataSource) MessageBuilder {
	c := make(chan []string)

	go func() {
		for {
			var lines []string

			scanner := bufio.NewScanner(strings.NewReader(<-source))
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					lines = append(lines, line)
				}
			}
			if scanner.Err() == nil && len(lines) > 0 {
				c <- lines
			}
		}
	}()

	return c
}

func WordGroups(source DataSource) MessageBuilder {
	c := make(chan []string)

	go func() {
		for {
			var words []string

			scanner := bufio.NewScanner(strings.NewReader(<-source))
			scanner.Split(bufio.ScanWords)
			for scanner.Scan() {
				words = append(words, scanner.Text())
			}
			if scanner.Err() == nil && len(words) > 0 {
				// Combine into groups so that they're as big as they can
				// be without going over a limit.
				var groups []string
				group, words := words[0], words[1:]

				for _, word := range words {
					if len(group)+1+len(word) <= 20 {
						group = fmt.Sprintf("%s %s", group, word)
					} else {
						groups = append(groups, group)
						group = word
					}
				}
				if group != "" {
					groups = append(groups, group)
				}

				c <- groups
			}
		}
	}()

	return c
}
