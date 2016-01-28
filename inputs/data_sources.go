package inputs

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/subparlabs/bonjourno/log"
)

// DataSource - Provides data to be turned into messages
type DataSource <-chan string

func StaticText(text string) (DataSource, error) {
	c := make(chan string)

	go func() {
		c <- text
	}()

	return c, nil
}

func FileWatcher(filename string) (DataSource, error) {
	// Make sure we can open the file, even though it might go away later
	if f, err := os.Open(filename); err != nil {
		f.Close()
		return nil, err
	} else {
		f.Close()
	}

	c := make(chan string)
	go func() {
		var fileContents []byte

		for {
			if f, err := os.Open(filename); err == nil {
				defer f.Close()

				// Read and send if contents have changed
				if newContents, err := ioutil.ReadAll(f); err != nil {
					log.Error("Failed to read file", "err", err)
				} else if bytes.Compare(newContents, fileContents) != 0 {
					fileContents = newContents
					c <- string(fileContents)
				}
			}

			time.Sleep(3 * time.Second)
		}
	}()

	return c, nil
}

func Download(url string) (DataSource, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if len(body) == 0 {
		return nil, errors.New("Empty response from that url")
	}

	return StaticText(string(body))
}
