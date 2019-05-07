package liveview

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/andybalholm/escaper"
	"github.com/gorilla/websocket"
)

// An ElementRef points to an element in a live view, and can be used to
// perform actions on it remotely.
type ElementRef struct {
	c        *Controller
	v        View
	selector string
}

// QuerySelector returns a reference to the first element in the view that
// matches selector. It does not check to make sure that the element actually
// exists in the client browser.
func (c *Controller) QuerySelector(v View, selector string) *ElementRef {
	return &ElementRef{
		c:        c,
		v:        v,
		selector: selector,
	}
}

type scriptMessage struct {
	ID       string `json:"id"`
	Selector string `json:"selector"`
	Action   string `json:"action"`
}

// Do executes a snippet of JavaScript on the client. It runs with "this" set
// to the element referenced by e.
func (e *ElementRef) Do(script ...interface{}) error {
	ch := e.c.channelForView(e.v)
	if ch == nil {
		return errors.New("channel not found")
	}
	if ch.socket == nil {
		return fmt.Errorf("channel %s not connected yet", ch.id)
	}

	sm := scriptMessage{
		ID:       ch.id,
		Selector: e.selector,
		Action:   buildJS(script...),
	}
	p, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	return ch.socket.WriteMessage(websocket.TextMessage, p)
}

func buildHTML(args ...interface{}) string {
	var sb strings.Builder
	e := escaper.New(&sb)
	e.Print(args...)
	return sb.String()
}

func buildJS(args ...interface{}) string {
	s := buildHTML("<script>", escaper.List(args))
	return strings.TrimPrefix(s, "<script>")
}

// SetTextContent sets e's textContent to s.
func (e *ElementRef) SetTextContent(s string) error {
	return e.Do(`this.textContent = `, s)
}

// SetInnerHTML sets e's innerHTML. The content is processed with an Escaper from
// github.com/andybalholm/escaper.
func (e *ElementRef) SetInnerHTML(content ...interface{}) error {
	return e.Do(`this.innerHTML = `, buildHTML(content...))
}
