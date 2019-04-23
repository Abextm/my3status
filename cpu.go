package my3status

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CPUColors struct {
	User      string
	Nice      string
	System    string
	Idle      string
	IOWait    string
	IRQ       string
	SoftIRQ   string
	Steal     string
	Guest     string
	GuestNice string
	Other     string
}

type statSample struct {
	Stats []int64
	Time  time.Time
	Next  *statSample
}

type colorSegment struct {
	Color  string
	Shares int64
	Runes  int
}

type CPU struct {
	Colors *CPUColors

	// How many chars wide to be
	Width int

	// How long the immediate cpu counter should be, or zero for off
	ShortInterval time.Duration
	oldSample     *statSample
	newSample     *statSample

	Show1  bool
	Show5  bool
	Show15 bool

	stat    ProcFile
	loadavg ProcFile
}

func (c *CPU) Status() (StatusBlock, error) {
	minlen := c.Width
	segments := 0
	if c.ShortInterval != 0 {
		segments++
	}
	if c.Show1 {
		segments++
	}
	if c.Show5 {
		segments++
	}
	if c.Show15 {
		segments++
	}
	tml := (segments * 4) - 1
	if minlen < tml {
		minlen = tml
	}
	if segments == 0 {
		return StatusBlock{}, fmt.Errorf("CPU: enable Show{1,5,15} and/or set ShortInterval")
	}

	runes := make([]rune, 0, minlen)
	writtenSegs := 0
	front := (minlen - tml) / (segments * 2)
	runes = pad(runes, front, false)

	colorSegments := make([]colorSegment, 0, 10)
	totalColorShares := int64(0)

	if c.ShortInterval != 0 || c.Colors != nil {
		now := time.Now()
		newSample := &statSample{
			Time: now,
		}

		cpus := 0
		var times []int64
		{
			stat, err := c.stat.Read(`/proc/stat`)
			if err != nil {
				return StatusBlock{}, err
			}

			endCpuAll := bytes.IndexByte(stat, '\n')
			for remStat := stat[endCpuAll+1:]; len(remStat) != 0 && bytes.HasPrefix(remStat, []byte("cpu")); cpus++ {
				idx := bytes.IndexByte(remStat, '\n')
				if idx < 0 {
					break
				}
				remStat = remStat[idx+1:]
			}
			statLine := string(stat[5:endCpuAll])
			stats := strings.Split(statLine, " ")
			parts := make([]int64, len(stats))
			for i := range stats {
				parts[i], err = strconv.ParseInt(stats[i], 10, 64)
			}
			newSample.Stats = parts
		}

		if c.newSample != nil {
			c.newSample.Next = newSample
		}
		c.newSample = newSample

		oldTime := now.Add(-c.ShortInterval)
		if c.oldSample == nil {
			c.oldSample = c.newSample
		}
		s := c.oldSample
		for ; s != nil && s.Time.Before(oldTime); s = s.Next {
		}
		c.oldSample = s
		if times == nil {
			times = make([]int64, len(c.newSample.Stats))
		}
		for i := range times {
			times[i] = c.newSample.Stats[i] - c.oldSample.Stats[i]
		}

		allTime := int64(0)
		for _, s := range times {
			allTime += s
		}

		if c.ShortInterval != 0 {
			immLA := (float64(allTime-times[3]) * float64(cpus)) / float64(allTime)
			runes = append(runes, []rune(fmt.Sprintf("%.2f", immLA))...)
			writtenSegs++
			runes = pad(runes, front+(minlen*writtenSegs)/segments, true)
		}

		if c.Colors != nil {
			i := 0
			add := func(c string) {
				val := times[i]
				i++
				if c == "" {
					return
				}
				for s := range colorSegments {
					if colorSegments[s].Color == c {
						colorSegments[s].Shares += val
						goto next
					}
				}
				colorSegments = append(colorSegments, colorSegment{
					Color:  c,
					Shares: val,
				})
			next:
				totalColorShares += val
			}

			add(c.Colors.User)
			add(c.Colors.Nice)
			add(c.Colors.System)
			add(c.Colors.Idle)
			add(c.Colors.IOWait)
			add(c.Colors.IRQ)
			add(c.Colors.SoftIRQ)
			add(c.Colors.Steal)
			add(c.Colors.Guest)
			add(c.Colors.GuestNice)
		}
	}

	{
		loadavg, err := c.loadavg.Read(`/proc/loadavg`)
		if err != nil {
			return StatusBlock{}, err
		}

		las := strings.SplitN(string(loadavg), " ", 4)

		if c.Show1 {
			runes = append(runes, []rune(las[0])...)
			writtenSegs++
			runes = pad(runes, front+(minlen*writtenSegs)/segments, true)
		}

		if c.Show5 {
			runes = append(runes, []rune(las[1])...)
			writtenSegs++
			runes = pad(runes, front+(minlen*writtenSegs)/segments, true)
		}

		if c.Show15 {
			runes = append(runes, []rune(las[2])...)
			writtenSegs++
		}
	}
	runes = pad(runes, minlen, false)

	if totalColorShares <= 0 {
		return StatusBlock{
			FullText: string(runes),
		}, nil
	}

	segPtrs := make([]int, len(colorSegments))
	for i := range segPtrs {
		segPtrs[i] = i
	}

	sort.SliceStable(segPtrs, func(ii, ji int) bool {
		return colorSegments[segPtrs[ii]].Shares < colorSegments[segPtrs[ji]].Shares
	})

	unallocedRunes := int64(len(runes))
	for _, sp := range segPtrs {
		rs := colorSegments[sp].Shares * int64(len(runes)) / totalColorShares
		if rs <= 0 && colorSegments[sp].Shares > (totalColorShares/int64(len(runes)*2)) {
			rs = 1
		}
		if rs > unallocedRunes {
			rs = unallocedRunes
		}
		colorSegments[sp].Runes = int(rs)
		unallocedRunes -= rs
	}

	buf := &bytes.Buffer{}
	for _, seg := range colorSegments {
		if seg.Runes <= 0 {
			continue
		}
		fmt.Fprintf(buf, `<span %s>%s</span>`, seg.Color, string(runes[:seg.Runes]))
		runes = runes[seg.Runes:]
	}
	if len(runes) > 0 {
		fmt.Fprintf(buf, `<span %s>%s</span>`, c.Colors.Other, string(runes))
	}
	return StatusBlock{
		FullText: buf.String(),
		Markup:   MarkupPango,
	}, nil
}

func pad(arr []rune, count int, min bool) []rune {
	if min {
		arr = append(arr, ' ')
	}
	for len(arr) < count {
		arr = append(arr, ' ')
	}
	return arr
}

func HTOPAdvancedCPUColors() *CPUColors {
	return &CPUColors{
		User:      `underline="single" underline_color="#00FF00"`,
		Nice:      `underline="single" underline_color="#0000FF"`,
		System:    `underline="single" underline_color="#FF0000"`,
		IOWait:    `underline="single" underline_color="#7F7F7F"`,
		IRQ:       `underline="single" underline_color="#FFAE00"`,
		SoftIRQ:   `underline="single" underline_color="#FF60A0"`,
		Steal:     `underline="single" underline_color="#000000"`,
		Guest:     `underline="single" underline_color="#00FFFF"`,
		GuestNice: `underline="single" underline_color="#007FFF"`,
		Other:     `underline="single" underline_color="#FFFFFF"`,
	}
}
