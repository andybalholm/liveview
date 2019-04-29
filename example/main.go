package main

import (
	"io"
	"net/http"
	"time"

	"github.com/andybalholm/escaper"
	"github.com/andybalholm/liveview"
)

var lvc = new(liveview.Controller)

func main() {
	http.Handle("/live-view/", lvc)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		e := escaper.New(w)
		e.Print(
			`<!doctype html>
			<html>
			  <head>
				<title>Live View Example</title>
			  </head>

			  <body>
				<h1><a href="/">Home</a></h1>`,
		)
		lvc.Render(e, new(ClickCounter))
		lvc.Render(e, new(CurrentTime))
		lvc.Render(e, new(CheckboxExample))
		lvc.Render(e, new(TextboxExample))
		e.Print(liveview.JSTag)
		e.Print(`</body></html>`)
	})

	http.ListenAndServe(":8000", nil)
}

type ClickCounter struct {
	count int
}

func (c *ClickCounter) Render(w io.Writer) {
	e := escaper.New(w)
	e.Print(
		`<button live-click="decrement">-</button>`,
		c.count,
		`<button live-click="increment">+</button>`,
	)
}

func (c *ClickCounter) HandleEvent(evt liveview.Event) {
	switch evt.Event {
	case "increment":
		c.count++
		lvc.Update(c)
	case "decrement":
		c.count--
		lvc.Update(c)
	}
}

type CurrentTime struct {
	_ byte
}

func (CurrentTime) Render(w io.Writer) {
	e := escaper.New(w)
	e.Print(`<time>`, time.Now().Format("Jan 2, 2006 3:04:05 PM"), `</time>`)
}

func (c *CurrentTime) HandleEvent(evt liveview.Event) {
	switch evt.Event {
	case "connect":
		go func() {
			for {
				err := lvc.Update(c)
				if err != nil {
					return
				}
				time.Sleep(time.Second)
			}
		}()
	}
}

type CheckboxExample struct {
	checked bool
}

func (c *CheckboxExample) Render(w io.Writer) {
	e := escaper.New(w)
	e.Print(`<label><input type="checkbox" live-change="toggle" `)
	if c.checked {
		e.Print(`checked`)
	}
	e.Print(
		`>`,
		c.checked,
		`</label>`,
	)
}

func (c *CheckboxExample) HandleEvent(evt liveview.Event) {
	switch evt.Event {
	case "toggle":
		c.checked = evt.Value == "true"
		lvc.Update(c)
	}
}

type TextboxExample struct {
	value string
}

func (t *TextboxExample) Render(w io.Writer) {
	e := escaper.New(w)
	e.Print(
		`<label>Echo your input: <input live-input=input value="`, t.value, `"></label>
		<div>`, t.value, `</div>`,
	)
}

func (t *TextboxExample) HandleEvent(evt liveview.Event) {
	switch evt.Event {
	case "input":
		t.value = evt.Value
		lvc.Update(t)
	}
}
