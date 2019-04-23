package my3status

import "image/color"

// Alignment controls the direction to float the text
// AlignLeft is the default
type Alignment string

// Alignment controls the direction to float the text
// AlignLeft is the default
const (
	AlignLeft   Alignment = "left"
	AlignCenter Alignment = "center"
	AlignRight  Alignment = "right"
)

// Markup indicates how the text of the block should be parsed.
type Markup string

const (
	// MarkupNone indicates the string will be used as-is
	MarkupNone Markup = "none"

	// MarkupPango markup only works if you use a pango font.
	MarkupPango Markup = "pango"
)

// StatusBlock is a block that is rendered by i3bar
type StatusBlock struct {
	// The FullText will be displayed by i3bar on the status line.
	// This is the only required key. If full_text is an empty string,
	// the block will be skipped.
	FullText string

	// ShortText will be used in case the status line needs to be shortened
	// because it uses more space than your screen provides. For example, when
	// displaying an IPv6 address, the prefix is usually (!) more relevant than
	// the suffix, because the latter stays constant when using autoconf, while
	// the prefix changes. When displaying the date, the time is more important
	// than the date (it is more likely that you know which day it is than what
	// time it is).
	ShortText string

	// To make the current state of the information easy to spot, colors can be
	// used. For example, the wireless block could be displayed in red if the
	// card is not associated with any network and in green or yellow (depending
	// on the signal strength) when it is associated.
	Color color.Color

	// Background overrides the background color for this particular block.
	Background color.Color

	// Overrides the border color for this particular block.
	Border color.Color

	// The minimum width (in pixels) of the block. If the content of the FullText
	// field take less space than the specified min_width, the block will be
	// padded to the left and/or the right side, according to the align key. This
	// is useful when you want to prevent the whole status line to shift when
	//value take more or less space between each iteration. The value can also be
	// a string. In this case, the width of the text given by min_width determines
	// the minimum width of the block. This is useful when you want to set a
	// sensible minimum width regardless of which font you are using, and at what
	// particular size.
	MinWidth int

	// Align text on the center, right or left (default) of the block, when the
	// minimum width of the latter, specified by the min_width key, is not reached.
	Align Alignment

	// Urgent specifies whether the current value is urgent. Examples are battery
	// charge values below 1 percent or no more available disk space (for non-root
	// users). The presentation of urgency is up to i3bar.
	Urgent bool

	// Markup indicates how the text of the block should be parsed.
	// Pango markup only works if you use a pango font.
	Markup Markup

	// Separator optionally controls the space after this Widget. If unset the
	// DefaultSeparator will be used
	Separator

	Extra map[string]interface{}
}

func (s StatusBlock) Status() (StatusBlock, error) {
	return s, nil
}

type Separator struct {
	// Hide specifies whether a separator line should be drawn after this block.
	// The default is false, meaning the separator line will be drawn. Note that
	// if you disable the separator line, there will still be a gap after the
	// block, unless you also use Width.
	Hide *bool

	// Width controls the amount of pixels to leave blank after the block. In the
	// middle of this gap, a separator line will be drawn unless separator is
	// disabled. Normally, you want to set this to an odd value (the default is
	// 9 pixels), since the separator line is drawn in the middle.
	Width *int
}

// Widget provides StatusBlocks to the i3bar
type Widget interface {
	Status() (StatusBlock, error)
}

// A ClickEvent is fired when the user interacts with a specific ClickableWidget
type ClickEvent struct {
	// X11 root window coordinates where the click occurred
	X int `json:"x"`
	Y int `json:"y"`

	// X11 button ID (for example 1 to 3 for left/middle/right mouse button)
	Button int `json:"button"`

	// Coordinates where the click occurred, with respect to the top left corner
	// of the block
	RelativeX int `json:"relative_x"`
	RelativeY int `json:"relative_y"`

	// Width and height (in px) of the block
	Width  int `json:"width"`
	Height int `json:"height"`

	// An array of the modifiers active when the click occurred. The order in
	// which modifiers are listed is not guaranteed.
	Modifiers []string `json:"modifiers"`
}

// A ClickableWidget is a Widget that can receive ClickEvents
type ClickableWidget interface {
	Widget
	Click(ClickEvent) bool
}
