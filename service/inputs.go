package service

import (
	"bufio"
	"errors"
	"os"
	"time"

	"github.com/subparlabs/bonjourno/log"
)

type InputStream interface {
	Get() string
}

type PriorityMultistream struct {
	streams []InputStream
}

func NewPriorityMultistream(streams []InputStream) (InputStream, error) {
	// Filter out nil streams. This method does it in-place, using same
	// underlying array, YAY!
	realStreams := streams[:0]
	for _, stream := range streams {
		if stream != nil {
			realStreams = append(realStreams, stream)
		}
	}

	if len(realStreams) == 0 {
		return nil, errors.New("No input streams defined")
	} else if len(realStreams) == 1 {
		return realStreams[0], nil
	}

	return &PriorityMultistream{
		streams: realStreams,
	}, nil
}

func (pms *PriorityMultistream) Get() string {
	for _, stream := range pms.streams {
		if msg := stream.Get(); msg != "" {
			return msg
		}
	}

	return ""
}

type StaticText struct {
	text string
}

func NewStaticText(text string) (*StaticText, error) {
	if text == "" {
		return nil, errors.New("No text given")
	}

	return &StaticText{text}, nil
}

func (st StaticText) Get() string {
	return st.text
}

type FileLines struct {
	// Track position in lines with an index, so we can compare with a reload
	// of the file, and not loose state when nothing actually changed.
	lines []string
	index int

	filename string

	rotateTicker <-chan time.Time
	updateTicker <-chan time.Time
}

func NewFileLines(filename string, interval time.Duration) (*FileLines, error) {
	fl := &FileLines{
		rotateTicker: time.Tick(interval),
		updateTicker: time.Tick(time.Second * 3),
		filename:     filename,
	}

	// Init with lines right away, so basic errors, like file not existing
	// are caught when the user is paying attention.
	if err := fl.updateLines(); err != nil {
		return nil, err
	}

	return fl, nil
}

func (fl *FileLines) updateLines() error {
	f, err := os.Open(fl.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Save for the end, so we don't replace existing lines on error
	fl.lines = lines
	fl.index = fl.index % len(fl.lines)
	return nil
}

func (fl *FileLines) Get() string {
	select {
	case <-fl.rotateTicker:
		fl.index = (fl.index + 1) % len(fl.lines)
	case <-fl.updateTicker:
		// Update in Get() cuz it doesn't matter otherwise, and it's an easier
		// cleanup than keeping a goroutine running.
		if err := fl.updateLines(); err != nil {
			log.Error("Failed to update lines from file", "err", err)
		}
	default:
	}

	return fl.lines[fl.index]
}

type FileWatcher struct {
	filename string
}

func NewFileWatcher(filename string) (*FileWatcher, error) {
	fw := &FileWatcher{
		filename: filename,
	}

	// Make sure we can open the file, even though it might go away later
	if f, err := os.Open(fw.filename); err != nil {
		f.Close()
		return nil, err
	} else {
		f.Close()
	}

	return fw, nil
}

func (fw *FileWatcher) Get() string {
	f, err := os.Open(fw.filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			return line
		}
	}

	return ""
}
