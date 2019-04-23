package my3status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Config struct {
	// Widgets contains all Widgets to be rendered, left to right
	Widgets []Widget

	// The DefaultSeparator is used when a Widget does not provide it's
	// own Separator
	DefaultSeparator Separator

	// If false, watch the binary for changes and restart if it is changed. This
	// restarts the binary in place, and can leak fds opened by other packages.
	DontWatchBinary bool

	// How often to update the bar. If unset 1 second is used
	Interval time.Duration
}

const (
	envContinue  = "MY3STATUS_CONTINUE"
	envSeenToken = "MY3STATUS_SEEN_TOKEN"
	envValueYes  = "YES"
)

func (c Config) Loop() {
	cont, _ := os.LookupEnv(envContinue)
	isContinue := cont == envValueYes
	if !isContinue {
		_, err := os.Stdout.WriteString(`{"version": 1, "click_events": true}[`)
		if err != nil {
			panic(fmt.Errorf("unable to write header: %v", err))
		}
		os.Setenv(envContinue, envValueYes)
	}

	var mtime time.Time
	var binary string
	if !c.DontWatchBinary {
		var err error
		binary, err = os.Executable()
		if err != nil {
			panic(fmt.Errorf("unable to get executable: %v", err))
		}
		stat, err := os.Stat(binary)
		if err != nil {
			panic(fmt.Errorf("unable to stat executable: %v", err))
		}
		mtime = stat.ModTime()
	}

	out := make([]map[string]interface{}, 0, len(c.Widgets))
	enc := json.NewEncoder(os.Stdout)

	click := make(chan struct{})
	go func() {
		in := io.Reader(os.Stdin)
		hasTokenStr, _ := os.LookupEnv(envSeenToken)
		hasToken := hasTokenStr == envValueYes
		if hasToken {
			in = io.MultiReader(bytes.NewBufferString(`[{}`), in)
		}
		dec := json.NewDecoder(in)
		t, err := dec.Token()
		if err != nil {
			panic(fmt.Errorf("unable to read click header: %v", err))
		}
		if t != json.Delim('[') {
			panic(fmt.Errorf("got unexpected token waiting for header: %v", t))
		}
		os.Setenv(envSeenToken, envValueYes)

		for {
			data := struct {
				ClickEvent
				Name     string `json:"name"`
				Instance string `json:"instance"`
			}{}
			err := dec.Decode(&data)
			if err != nil {
				panic(fmt.Errorf("unable to read click event: %v", err))
			}
			index, err := strconv.Atoi(data.Instance)
			if err != nil || index < 1 || index > len(c.Widgets) {
				continue
			}
			widget := c.Widgets[index-1]
			cw, ok := widget.(ClickableWidget)
			if !ok {
				continue
			}
			redraw := cw.Click(data.ClickEvent)
			if redraw {
				click <- struct{}{}
			}
		}
	}()

	interval := c.Interval
	if interval == 0 {
		interval = time.Second
	}
	tick := time.NewTicker(interval).C
	for {
		if binary != "" {
			stat, err := os.Stat(binary)
			if err == nil {
				if stat.ModTime() != mtime {
					os.Stderr.WriteString("restarting\n")
					Restart()
				}
			}
		}

		for index, seg := range c.Widgets {
			s, err := seg.Status()

			value := make(map[string]interface{}, 16)

			if err != nil {
				s = StatusBlock{
					FullText: fmt.Sprintf("error: %v", err),
					Urgent:   true,
				}
			}

			for k, v := range s.Extra {
				value[k] = v
			}

			value["name"] = reflect.TypeOf(seg).String()
			value["instance"] = strconv.Itoa(index + 1)

			if c.DefaultSeparator.Hide != nil {
				value["separator"] = !*c.DefaultSeparator.Hide
			}
			if c.DefaultSeparator.Width != nil {
				value["separator_block_width"] = *c.DefaultSeparator.Width
			}
			if s.Separator.Hide != nil {
				value["separator"] = !*s.Separator.Hide
			}
			if s.Separator.Width != nil {
				value["separator_block_width"] = *s.Separator.Width
			}

			value["full_text"] = s.FullText
			if s.ShortText != "" {
				value["short_text"] = s.ShortText
			}
			encodeColor(value, "color", s.Color)
			encodeColor(value, "background", s.Background)
			encodeColor(value, "border", s.Border)
			if s.MinWidth != 0 {
				value["min_width"] = s.MinWidth
			}
			if s.Align != "" && s.Align != AlignLeft {
				value["align"] = string(s.Align)
			}
			if s.Urgent {
				value["urgent"] = true
			}
			if s.Markup != "" && s.Markup != MarkupNone {
				value["markup"] = string(s.Markup)
			}

			out = append(out, value)
		}
		err := enc.Encode(out)
		if err != nil {
			panic(fmt.Errorf("unable to write output: %v", err))
		}
		_, err = os.Stdout.WriteString(",\n")
		if err != nil {
			panic(fmt.Errorf("unable to write output: %v", err))
		}
		out = out[:0]

		select {
		case <-tick:
		case <-click:
		}
	}
}

func encodeColor(value map[string]interface{}, key string, c color.Color) {
	if c == nil {
		return
	}
	r, g, b, _ := c.RGBA()
	value[key] = fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}
