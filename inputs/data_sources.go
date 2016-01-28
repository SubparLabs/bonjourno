package inputs

import (
	"bytes"
	"os"
	"time"
)

// DataSource - Provides data to be turned into messages
type DataSource <-chan string

func StaticText(text string) (DataSource, error) {
	c := make(chan string)

	go func() {
		c <- text
		close(c)
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
				buffer := make([]byte, 100*1024*1024) // 100mb
				if numRead, err := f.Read(buffer); err == nil && bytes.Compare(buffer[:numRead], fileContents) != 0 {
					fileContents = buffer[:numRead]
					c <- string(fileContents)
				}
			}

			time.Sleep(3 * time.Second)
		}
	}()

	return c, nil
}
