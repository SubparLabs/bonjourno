package inputs

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// MessageFilter - Filter or modifies messages
type MessageFilter func(<-chan string) <-chan string

func LowerCase(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		for {
			out <- strings.ToLower(<-in)
		}
	}()

	return out
}

func UpperCase(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		for {
			out <- strings.ToUpper(<-in)
		}
	}()

	return out
}

func MixedCase(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		for {
			msg := <-in

			var mixed []rune
			for _, r := range msg {
				if rand.Intn(2) == 0 {
					mixed = append(mixed, unicode.ToUpper(r))
				} else {
					mixed = append(mixed, unicode.ToLower(r))
				}
			}

			out <- string(mixed)
		}
	}()

	return out
}

func LeetSpeak(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {

		// Sadly a lot of symbols don't work, so use unicode stuff
		leetMap := map[rune]rune{
			'a': '^', 'b': 'β', 'c': '¢', 'd': 'Ð', 'e': 'ë', 'f': 'ƒ',
			'g': '9', 'i': '1', 'l': '£', 'n': 'η', 'o': '¤', 'r': '®',
			's': '§', 't': '†', 'u': 'µ', 'v': '√', 'w': 'ω', 'x': 'Ж',
			'y': 'Ψ',
		}

		for {
			msg := <-in

			var leet []rune
			for _, r := range msg {
				if replace, ok := leetMap[unicode.ToLower(r)]; !ok || rand.Intn(2) == 0 {
					leet = append(leet, r)
				} else {
					leet = append(leet, replace)
				}
			}

			out <- string(leet)
		}
	}()

	return out
}

func RateLimit(interval time.Duration, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		for {
			out <- (<-in)
			time.Sleep(interval)
		}
	}()

	return out
}

func Prefix(prefix string, in <-chan string) <-chan string {
	if prefix == "" {
		return in
	}

	out := make(chan string)

	go func() {
		for {
			// Only apply prefix if it's a non-empty msg
			if msg := <-in; msg != "" {
				out <- prefix + msg
			} else {
				out <- ""
			}
		}
	}()

	return out
}

func Cleanup(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		endsRe := regexp.MustCompile("^[^a-zA-Z0-9-_|]+|[^a-zA-Z0-9-_|]+$")
		middleRe := regexp.MustCompile("[^a-zA-Z0-9-_|]+")

		for {
			msg := <-in

			// Some characters cause the service to be ignored completely.
			// Not sure which, so make a conservative conversion.
			// TODO: look up the spec and only replace actually invalid chars

			// Just remove stuff at the start & end. This also serves to trim
			msg = endsRe.ReplaceAllString(msg, "")

			// Replace multiple invalid chars in middle with a single -
			msg = middleRe.ReplaceAllString(msg, "-")

			// The Finder sidebar cuts off somewhere under 20, maybe less, but
			// browsing to the share in "Network" shows somewhere around 40.
			if len(msg) > 40 {
				msg = msg[:40]
			}

			out <- msg
		}
	}()

	return out
}

func LimitSize(size int, in <-chan string) <-chan string {
	if size <= 0 {
		return in
	}

	out := make(chan string)

	go func() {
		for {
			msg := <-in

			if len(msg) > size {
				msg = msg[:size]
			}

			out <- msg
		}
	}()

	return out
}
