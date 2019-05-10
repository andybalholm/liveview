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
    var { id, render, selector, action } = JSON.parse(data);

    var view = document.querySelector('[data-live-view="' + id + '"]');

	if (render) {
		morphdom(view.children[0], '<div>' + render + '</div>');
	}

	if (action) {
		var f = new Function(action);
		f.call(view.querySelector(selector));
	}
  });

  live_view.addEventListener('close', event => {
    // Do we need to do anything here?
  });

  // The form_params function is borrowed from https://github.com/grych/drab.
  // (c) 2016 Tomek "Grych" Gryszkiewicz
  function form_params(form) {
	  var params = {};
	  function add(key, value) {
		  if (params[key]) {
			  params[key].push(value);
		  } else {
			  params[key] = [value];
		  }
	  }
	  var inputs = form.querySelectorAll("input, textarea, select");
	  for (var i = 0; i < inputs.length; i++) {
		var input = inputs[i];
		var key = input.name || input.id || false;
		if (key) {
		  if (input.type == "radio" || input.type == 'checkbox') {
			if (input.checked) {
				add(key, input.value);
			}
		  } else if (input.type == "select-multiple") {
			  for (var j = 0; j < input.options.length; j++) {
				  var option = input.options[j];
				  if (option.selected) {
					  add(key, option.value);
				  }
			  }
		  } else if (input.type == "submit") {
			  // Ignore it.
		  } else {
			  add(key, input.value);
		  }
		}
	  };
	  return params;
  }

  [
    'click',
    'change',
    'input',
	'submit',
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
				if (typeof(target.value) == "string") {
					data.value = target.value;
				} else if (element.hasAttribute("live-value")) {
					data.value = element.getAttribute("live-value");
				}
				break;
			}

			var form;
			if (target.tagName == "FORM") {
				form = target;
			} else if (element.tagName == "FORM") {
				form = element;
			} else if (target.form) {
				form = target.form;
			}
			if (form) {
				data.form_data = form_params(form);
			}

			live_view.send(JSON.stringify(data));
		};

		var debounce = element.getAttribute("live-debounce");
		if (debounce) {
			clearTimeout(element.liveview_timeout);
			element.liveview_timeout = setTimeout(
				() => {
					element.liveview_timeout = null;
					send_event();
				},
				debounce
			);
		} else {
			send_event();
		}
      }
    });
  });
})();
`)
