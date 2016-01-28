package inputs

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/subparlabs/bonjourno/log"
)

// MessageBuilder - Turns raw data into a list of messages
type MessageBuilder <-chan []string

func CSVField(fieldIndex int, source DataSource) MessageBuilder {
	c := make(chan []string)

	go func() {
		for {
			reader := csv.NewReader(strings.NewReader(<-source))
			reader.TrimLeadingSpace = true
			reader.FieldsPerRecord = -1

			var values []string

			// Ignore first line - comment
			record, err := reader.Read()
			for err == nil {
				record, err = reader.Read()
				if err == nil && len(record) > fieldIndex {
					values = append(values, record[fieldIndex])
				}
			}
			if err != nil && err != io.EOF {
				log.Error("Failed to read CSV", "err", err)
			} else if len(values) > 0 {
				c <- values
			}
		}
	}()

	return c
}

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
