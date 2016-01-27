package service

import (
	"bufio"
	"errors"
	"os"
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
	lines []string
}

func NewFileLines(filename string) (*FileLines, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
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
		return nil, err
	} else if len(lines) == 0 {
		return nil, errors.New("No lines found in file")
	}

	return &FileLines{
		lines: lines,
	}, nil
}

func (fl *FileLines) Get() string {
	line := fl.lines[0]
	fl.lines = append(fl.lines[1:], fl.lines[0])
	return line
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
