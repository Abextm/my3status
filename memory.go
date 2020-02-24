package my3status

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Memory displays the amount of Memory Available/Total in gb
type Memory struct {
	meminfo ProcFile
}

func (m *Memory) Status() (StatusBlock, error) {
	meminfo, err := m.meminfo.Read(`/proc/meminfo`)
	if err != nil {
		return StatusBlock{}, err
	}

	total, err := readMemInfoLine(meminfo, "MemTotal")
	if err != nil {
		return StatusBlock{}, err
	}

	free, err := readMemInfoLine(meminfo, "MemAvailable")
	if err != nil {
		return StatusBlock{}, err
	}

	used := total - free

	const gib = 1024 * 1024 * 1024
	return StatusBlock{
		FullText: fmt.Sprintf("%.1f/%.1fG", float64(used)/gib, float64(total)/gib),
	}, nil
}

func readMemInfoLine(meminfo []byte, name string) (uint64, error) {
	name += ":"
	idx := bytes.Index(meminfo, []byte(name))
	if idx == -1 {
		return 0, fmt.Errorf("Memory: unable to find %q block", name)
	}
	line := meminfo[idx+len(name):]
	end := bytes.IndexByte(line, '\n')
	if end != -1 {
		line = line[:end]
	}
	start := bytes.IndexAny(line, "0123456789")
	if start == -1 {
		return 0, fmt.Errorf("Memory: no numbers in %q block", name)
	}
	line = line[start:]
	space := bytes.IndexByte(line, ' ')
	size := uint64(1)
	if space != -1 {
		end := line[space+1:]
		line = line[:space]
		switch strings.ToLower(string(end)) {
		case "gib", "gb":
			size *= 1024
			fallthrough
		case "mib", "mb":
			size *= 1024
			fallthrough
		case "kib", "kb":
			size *= 1024
		}
	}
	val, err := strconv.ParseUint(string(line), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Memory: err parsing %q block: %v", name, err)
	}
	return val * size, nil
}
