package main

import (
	"time"

	. "github.com/abextm/my3status"
)

func main() {
	RedirectStderr("/home/abex/.my3status")

	Config{
		Widgets: []Widget{
			&APCUPSDStatus{
				Host:     "10.0.0.2:3551",
				Interval: time.Second * 3,
			},
			&NvidiaTemperature{
				Format: "%s°G",
			},
			&CPU{
				Colors:        HTOPAdvancedCPUColors(),
				ShortInterval: time.Second * 5,
				Show1:         true,
				Show15:        true,
				Width:         24,
			},
			&Temperature{
				Path:    "/sys/class/hwmon/hwmon0/temp1_input",
				Divisor: 1000,
			},
			&Memory{},
			&Edit{
				Widget: Switcher{
					&Time{
						Format: `Monday January 01/02/2006 15:04:05`,
					},
					&Time{
						Format:       `Mon 15:04 MST`,
						LocationName: "UTC",
					},
					&Time{
						Format:       `Mon 15:04 MST`,
						LocationName: "Asia/Tokyo",
					},
				},
				Func: func(s *StatusBlock) {
					s.Separator.Width = IntPtr(8)
				},
			},
			StatusBlock{}, // to make the previous separator apply
		},
		DefaultSeparator: Separator{
			Hide:  BoolPtr(true),
			Width: IntPtr(24),
		},
	}.Loop()
}
