package my3status

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// The FuncWidget type is an adapter to use a normal function as a
// non-ClickableWidget
type FuncWidget func() (StatusBlock, error)

func (w FuncWidget) Status() (StatusBlock, error) {
	return w()
}

// Edit allows you to override fields from a child Widget's StatusBlocks
// Widget may be a ClickableWidget
type Edit struct {
	Widget Widget
	Func   func(*StatusBlock)
}

func (e *Edit) Status() (StatusBlock, error) {
	sb, err := e.Widget.Status()
	if err == nil {
		e.Func(&sb)
	}
	return sb, err
}

func (e *Edit) Click(c ClickEvent) bool {
	cw, ok := e.Widget.(ClickableWidget)
	if ok {
		return cw.Click(c)
	}
	return false
}

// BoolPtr returns &value
func BoolPtr(value bool) *bool {
	return &value
}

// IntPtr returns &value
func IntPtr(value int) *int {
	return &value
}

// RedirectStderr makes all writes to stderr go to the passed file. The file is
// always opened in append mode
func RedirectStderr(filename string) {
	fi, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0770)
	if err != nil {
		panic(fmt.Errorf("RedirectStderr: unable to open file %q: %v", filename, err))
	}

	err = unix.Dup2(int(fi.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		panic(fmt.Errorf("RedirectStderr: unable to redirect: %v", err))
	}
}
