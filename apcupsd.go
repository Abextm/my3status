package my3status

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type APCUPSDStatus struct {
	Host             string
	Interval         time.Duration
	lastStatus       StatusBlock
	lastStatusExpiry time.Time
	conn             net.Conn
}

func (a *APCUPSDStatus) Status() (StatusBlock, error) {
	if a.lastStatusExpiry.After(time.Now()) {
		return a.lastStatus, nil
	}

	if a.conn == nil {
		conn, err := net.DialTimeout("tcp", a.Host, time.Second)
		if err != nil {
			return StatusBlock{}, err
		}
		a.conn = conn
	}

	a.conn.SetDeadline(time.Now().Add(time.Second))

	err := a.write([]byte("status"))
	if err != nil {
		return StatusBlock{}, err
	}

	lines := map[string]string{}
	for {
		bline, err := a.read()
		if err != nil {
			return StatusBlock{}, err
		}
		if len(bline) == 0 {
			break
		}
		line := strings.SplitN(string(bline), ":", 2)
		key := strings.TrimSpace(line[0])
		val := strings.TrimSpace(line[1])
		lines[key] = val
	}

	nompwr, err := parse(lines["NOMPOWER"])
	if err != nil {
		return StatusBlock{}, err
	}
	loadpct, err := parse(lines["LOADPCT"])
	if err != nil {
		return StatusBlock{}, err
	}
	load := fmt.Sprintf("%.0fW", (loadpct/100)*nompwr)
	status := StatusBlock{
		FullText:  load,
		ShortText: load,
	}

	if lines["STATUS"] != "ONLINE" {
		if lines["STATUS"] == "ONBATT" {
			status.Background = color.RGBA{R: 0xFF}
		} else {
			status.Background = color.RGBA{R: 0xFF, G: 0xFF, B: 0x00}
		}

		timeleft, _ := parse(lines["TIMELEFT"])
		remaining := fmt.Sprintf(" (%.1f Min)", timeleft)
		status.ShortText += remaining
		status.FullText += remaining
	}

	a.lastStatus = status
	a.lastStatusExpiry = time.Now().Add(a.Interval)
	return a.lastStatus, nil
}

var suffixes = []string{
	" Minutes",
	" Seconds",
	" Percent",
	" Volts",
	" Watts",
	" Hz",
	" C",
}

func parse(s string) (float64, error) {
	for _, suf := range suffixes {
		s = strings.TrimSuffix(s, suf)
	}

	return strconv.ParseFloat(s, 64)
}

func (a *APCUPSDStatus) write(data []byte) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(len(data)))
	_, err := a.conn.Write(buf)
	if err != nil {
		return err
	}
	_, err = a.conn.Write(data)
	return err
}

func (a *APCUPSDStatus) read() ([]byte, error) {
	buf := make([]byte, 2)
	_, err := io.ReadFull(a.conn, buf)
	if err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint16(buf)
	buf = make([]byte, size)
	_, err = io.ReadFull(a.conn, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
