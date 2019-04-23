package main

import (
	"time"

	. "github.com/abextm/my3status"
)

func main() {
	RedirectStderr("/home/abex/.my3status")

	Config{
		Widgets: []Widget{
			&Temperature{
				Path:    "/sys/class/drm/card0/device/hwmon/hwmon*/temp1_input",
				Divisor: 1000,
				Format:  "%.f°G",
			},
			&CPU{
				Colors:        HTOPAdvancedCPUColors(),
				ShortInterval: time.Second * 5,
				Show1:         true,
				Show15:        true,
				Width:         24,
			},
			&Temperature{
				Path:    "/sys/devices/platform/it87.552/hwmon/hwmon*/temp1_input",
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
