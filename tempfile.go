package my3status

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
)

// Reads a file off disk, parses it as an int, then formats that
// Typically this is used for temperature sensors in /sys/
type Temperature struct {
	Path    string
	Divisor float64
	Format  string
	file    ProcFile
}

func (t *Temperature) Status() (StatusBlock, error) {
	paths, err := filepath.Glob(t.Path)
	if err != nil {
		return StatusBlock{}, fmt.Errorf("Temp: %v", err)
	}
	if len(paths) != 1 {
		return StatusBlock{}, fmt.Errorf("Temp: %q does not match one file: %v", t.Path, paths)
	}
	contents, err := t.file.Read(paths[0])
	if err != nil {
		return StatusBlock{}, err
	}

	contents = bytes.TrimRight(contents, "\n")

	val, err := strconv.Atoi(string(contents))
	if err != nil {
		return StatusBlock{}, fmt.Errorf("Temp: %q contains non integer data: %v", contents, err)
	}

	value := float64(val)
	if t.Divisor != 0 {
		value /= t.Divisor
	}

	format := t.Format
	if format == "" {
		format = "%.0f°C"
	}

	return StatusBlock{
		FullText: fmt.Sprintf(format, value),
	}, nil
}
