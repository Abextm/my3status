package my3status

import (
	"time"
)

// Time is a Widget that renders a Clock
type Time struct {
	// Full width format specifier; see time.Format
	Format      string
	ShortFormat string

	// Name of the timezone to use
	LocationName string
	Location     *time.Location
}

func (t *Time) Status() (StatusBlock, error) {
	now := time.Now()

	if t.Location == nil && t.LocationName != "" {
		loc, err := time.LoadLocation(t.LocationName)
		if err != nil {
			return StatusBlock{}, err
		}
		t.Location = loc
	}

	if t.Location != nil {
		now = now.In(t.Location)
	}

	block := StatusBlock{}
	block.FullText = now.Format(t.Format)
	if t.ShortFormat != "" {
		block.ShortText = now.Format(t.ShortFormat)
	}
	return block, nil
}
