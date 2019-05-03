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
	  var target = event.target;
      var element = target.closest('[live-' + event_type + ']');

      if(element) {
        var event_name = element.getAttribute('live-' + event_type);
		event.preventDefault();
        var channel = element
          .closest('[data-live-view]')
          .getAttribute('data-live-view')

		var send_event = () => {
			var data = {
				event: event_name,
				channel: channel,
			};
			switch(element.type) {
			  case "checkbox":
				data.value = target.checked + "";
				break;
			  default:
				var value = target.value;
				if (typeof(value) == "string") {
					data.value = value;
				}
				break;
			}

			live_view.send(JSON.stringify(data));
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
