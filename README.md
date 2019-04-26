# Liveview

A Go package for building server-rendered HTML components with client-side interactivity.

These views use my escaper package (github.com/andybalholm/escaper) to render themselves to HTML.

## Usage

For a complete example program, see the `example` directory.

First you need to create a Controller, and configure your HTTP router to route request to it:

```go
var lvc = new(liveview.Controller)
http.Handle("/live-view/", lvc)
```

Each view needs to be implemented by a type that implements the `View` interface:

```go
// A View is a component that can render itself to HTML, and respond to Events.
type View interface {
	Render(e *escaper.Escaper)
	HandleEvent(e Event)
}
```

When you are rendering the initial page, insert a View with the Controller’s
`Render` method. You can have multiple views on one page, if you want.
You should also include script tags to load the JavaScript to make the live views work.

```go
lvc.Render(e, new(ClickCounter))
lvc.Render(e, new(CurrentTime))
e.Print(liveview.JSTag)
```

When a view needs to be refreshed, tell the Controller to update it.
It will re-render the view, and send the updated HTML to the client 
over a websocket.

```go
lvc.Update(v)
```

To define an event that will be passed to your view’s HandleEvent method,
give an HTML element an attribute starting with `live-` followed by the event name.
(The currently supported events are `click`, `change`, and `input`.)

```html
<button live-click="decrement">-</button>
```

When the user clicks on this button, the View will received a "decrement" Event.
