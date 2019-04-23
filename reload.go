package my3status

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

var callbackLock = &sync.Mutex{}
var restartCallbacks = []func(){}

// Restart causes the running application to be restarted in place
// This happens immediately after calling all callbacks added with
// BeforeRestart. Any goroutines that need synchronization MUST use
// BeforeRestart to create said synchronization
func Restart() {
	binary, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("Restart: unable to get executable: %v", err))
	}
	for _, cb := range restartCallbacks {
		cb()
	}
	err = unix.Exec(binary, os.Args, os.Environ())
	if err != nil {
		panic(fmt.Errorf("Restart: unable to execve new binary: %v", err))
	}
}

// BeforeRestart adds callback to a list of methods to be called before
// the application is restarted
func BeforeRestart(callback func()) {
	callbackLock.Lock()
	defer callbackLock.Unlock()
	restartCallbacks = append(restartCallbacks, callback)
}

// CloseFileBeforeRestart marks the file to be closed in case of a Restart
// This is logically similar to `BeforeRestart(func(){file.Close()}), though
// more efficient
func CloseFileBeforeRestart(file *os.File) {
	unix.CloseOnExec(int(file.Fd()))
}
