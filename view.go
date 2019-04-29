package liveview

import "io"

// An Event represents an action by the user, such as clicking a button. There
// are also special Events named "connect" and "disconnect" that take place
// when the View's websocked is connected and disconnected.
type Event struct {
	// Event is the name that was assigned to the event in the HTML markup
	// (e.g. in the live-click attribute).
	Event string `json:"event"`

	// If the target of the event is a form control, Value is its current value.
	Value string `json:"value"`

	ChannelID string `json:"channel"`
}

// A View is a component that can render itself to HTML, and respond to Events.
type View interface {
	Render(w io.Writer)
	HandleEvent(e Event)
}
