package my3status

import (
	"fmt"
	"io"
	"os"
)

type ProcFile struct {
	path string
	buf  []byte

	fi *os.File
}

func (p *ProcFile) Read(path string) ([]byte, error) {
	if p.path != path {
		p.path = path
		if p.fi != nil {
			p.fi.Close()
			p.fi = nil
		}
	}
	if p.fi == nil {
		fi, err := os.Open(p.path)
		if err != nil {
			return nil, fmt.Errorf("procfile: unable to open %q: %v", p.path, err)
		}
		CloseFileBeforeRestart(fi)
		p.fi = fi
	}

	_, err := p.fi.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, fmt.Errorf("procfile: unable to seek %q: %v", p.path, err)
	}

	if p.buf == nil {
		p.buf = make([]byte, 512)
	}

	buf := p.buf
	read := 0
	for {
		thisRead, err := io.ReadFull(p.fi, buf[read:cap(buf)])
		read += thisRead
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			p.buf = buf[:read]
			return p.buf, nil
		}
		if err != nil {
			return nil, err
		}
		oldbuf := buf[:read]
		buf = make([]byte, cap(buf)+512)
		copy(buf, oldbuf)
	}
}
