package my3status

import (
	"bufio"
	"fmt"
	"os/exec"
)

type NvidiaTemperature struct {
	Format string
	live   bool
	status StatusBlock
	err    error
}

func (t *NvidiaTemperature) Status() (StatusBlock, error) {
	if !t.live {
		t.live = true

		go func() {
			cmd := exec.Command("nvidia-smi", "--query-gpu=temperature.gpu", "--format=csv,noheader", "-l", "1")
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				t.err = err
				return
			}
			br := bufio.NewReader(pipe)
			err = cmd.Start()
			if err != nil {
				t.err = err
				return
			}
			for {
				line, err := br.ReadString('\n')
				if err != nil {
					t.err = err
					return
				}

				str := line[:len(line)-1]
				t.status.FullText = fmt.Sprintf(t.Format, str)
			}
		}()
	}

	return t.status, t.err
}
