package liveview

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/escaper"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
)

// A channel links a View, and ID, and a websocket.
type channel struct {
	view    View
	id      string
	socket  *synchronizedWebsocket
	created time.Time
}

// A Controller manages a collection of live views and their associated
// websockets.
type Controller struct {
	channels map[string]*channel
	byView   map[View]*channel
	lastGC   time.Time

	m sync.RWMutex
}

func (c *Controller) channelForView(v View) *channel {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.byView[v]
}

func (c *Controller) channelByID(id string) *channel {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.channels[id]
}

func (c *Controller) removeChannel(ch *channel) {
	c.m.Lock()
	defer c.m.Unlock()

	delete(c.channels, ch.id)
	delete(c.byView, ch.view)
}

// Register registers Views with c, and prepares to receive websocket
// connections for them.
func (c *Controller) Register(views ...View) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.channels == nil {
		c.channels = make(map[string]*channel)
		c.byView = make(map[View]*channel)
		c.lastGC = time.Now()
	} else {
		// Clean up any channels that have been waiting too long.
		gcThreshold := time.Now().Add(-30 * time.Second)
		if c.lastGC.Before(gcThreshold) {
			for id, ch := range c.channels {
				if ch.socket == nil && ch.created.Before(gcThreshold) {
					delete(c.channels, id)
					delete(c.byView, ch.view)
				}
			}
			c.lastGC = time.Now()
		}
	}

	for _, v := range views {
		ch := &channel{
			view:    v,
			created: time.Now(),
			id:      ksuid.New().String(),
		}

		c.channels[ch.id] = ch
		c.byView[v] = ch
	}
}

var upgrader = websocket.Upgrader{
	EnableCompression: true,
}

type subscription struct {
	Channel string `json:"subscribe"`
}

// ServeHTTP implements the http.Handler interface, to accept incoming
// websocket connections and serve the necessary JavaScript file. The
// Controller should be set up to handle the "/live-view/" path.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/live-view/socket":
		s, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		subscriptions := make(map[string]*channel)

		ss := &synchronizedWebsocket{Conn: s}

		for {
			_, p, err := s.ReadMessage()
			if err != nil {
				s.Close()
				break
			}

			var sub subscription
			err = json.Unmarshal(p, &sub)
			if err == nil && sub.Channel != "" {
				ch := c.channelByID(sub.Channel)
				if ch != nil && ch.socket == nil {
					subscriptions[sub.Channel] = ch
					ch.socket = ss
					ch.view.HandleEvent(Event{Event: "connect"})
				}
				continue
			}

			var evt Event
			err = json.Unmarshal(p, &evt)
			if err == nil && evt.Event != "" {
				ch := subscriptions[evt.ChannelID]
				if ch != nil {
					ch.view.HandleEvent(evt)
				}
				continue
			}
		}

		// The socket has shut down. Notify views, and clean up.
		for _, ch := range subscriptions {
			c.removeChannel(ch)
			ch.view.HandleEvent(Event{Event: "disconnect"})
		}

	case "/live-view/live-view.js":
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(liveViewJS)

	default:
		http.NotFound(w, r)
	}
}

// Render renders v to e, wrapped in a div that makes it a live view. The View
// is automatically Registered if it was not registered already.
func (c *Controller) Render(w io.Writer, v View) {
	ch := c.channelForView(v)
	if ch == nil {
		c.Register(v)
		ch = c.channelForView(v)
	}

	fmt.Fprintf(w, `<div data-live-view="%s"><div>`, ch.id)
	v.Render(w)
	io.WriteString(w, `</div></div>`)
}

// JSTag is the script tags that should be included in pages that use live
// views.
const JSTag template.HTML = `<script src="https://cdn.jsdelivr.net/gh/patrick-steele-idem/morphdom/dist/morphdom-umd.js"></script>
<script src="/live-view/live-view.js"></script>`

type update struct {
	Render string `json:"render"`
	ID     string `json:"id"`
}

// Update re-renders v and sends the updated HTML to the client.
func (c *Controller) Update(v View) error {
	ch := c.channelForView(v)
	if ch == nil {
		return errors.New("channel not found")
	}
	if ch.socket == nil {
		return fmt.Errorf("channel %s not connected yet", ch.id)
	}

	b := new(strings.Builder)
	e := escaper.New(b)
	v.Render(e)

	u := update{
		Render: b.String(),
		ID:     ch.id,
	}
	p, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return ch.socket.WriteMessage(websocket.TextMessage, p)
}

type synchronizedWebsocket struct {
	*websocket.Conn
	m sync.Mutex
}

func (s *synchronizedWebsocket) WriteMessage(messageType int, data []byte) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.Conn.WriteMessage(messageType, data)
}
