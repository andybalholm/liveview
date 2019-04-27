package liveview

var liveViewJS = []byte(
	`(() => {
  var url = new URL(location.href);
  url.protocol = url.protocol.replace('http', 'ws');
  url.pathname = '/live-view/socket';
  var live_view = new WebSocket(url);
  live_view.addEventListener('open', event => {
    document.querySelectorAll('[data-live-view]')
      .forEach(view => {
        live_view.send(JSON.stringify({
          subscribe: view.getAttribute('data-live-view'),
        }))
      });
  });

  live_view.addEventListener('message', event => {
    var data = event.data;
    var { id, render } = JSON.parse(data);

    document.querySelectorAll('[data-live-view="' + id + '"]')
      .forEach(view => {
		morphdom(view.children[0], '<div>' + render + '</div>');
      });
  });

  live_view.addEventListener('close', event => {
    // Do we need to do anything here?
  });

  [
    'click',
    'change',
    'input',
  ].forEach(event_type => {
    document.addEventListener(event_type, event => {
      var element = event.target;
      var event_name = element.getAttribute('live-' + event_type);

      if(typeof event_name === 'string') {
        var channel = event
          .target
          .closest('[data-live-view]')
          .getAttribute('data-live-view')

		var send_event = () => {
			var value = "";
			switch(element.type) {
			  case "checkbox":
				value = element.checked + "";
				break;
			  default:
				value = element.getAttribute('live-value') || element.value || "";
				break;
			}

			live_view.send(JSON.stringify({
			  event: event_name,
			  value: value,
			  channel: channel,
			}));
		};

		if (event_type == "input") {
			// Debounce it.
			clearTimeout(element.liveview_timeout);
			element.liveview_timeout = setTimeout(
				() => {
					element.liveview_timeout = null;
					send_event();
				},
				500
			);
		} else {
			send_event();
		}
      }
    });
  });
})();
`)
