function wsBackbone(url, start) {
	var ws = null;
  
	var subscriptions = {};
	var rpcCalls = {};

	window.wsRPC = function(meth, params, success) {
	  var id = Math.random().toString(36).substring(2);
		if (success != null) {
			rpcCalls[id] = success;
		}
    ws.sendIfReady(JSON.stringify({
		  Type: 'RPC',
			Method: {
			  Name: meth,
			  Id: id,
				Data: params,
			},
		}));
	};

	var closeSubscription = function(that) {
		var url = _.result(that, 'url') || urlError(); 
		if (subscriptions[url] != null) {
			logDebug('Unsubscribing from', url);
			ws.sendIfReady(JSON.stringify({
				Type: 'Unsubscribe',
				Object: {
					URI: url,
				},
			}));
			delete(subscriptions[url]);
		}
	};

  var backoff = 500;
  var setupWs = null;
	var state = {
	  reconnecting: false,
	  started: false,
		open: false,
	};
  setupWs = function() {
	  state.reconnecting = false;

		logInfo("Opening socket to", url);
	  ws = new WebSocket(url); 
		ws.sendIfReady = function(msg) {
			if (ws.readyState == 1) {
				ws.send(msg);
			} else {
				logError('Tried to send', msg, 'on', ws, 'in readyState', ws.readyState);
			}
		};
		ws.onclose = function(code, reason, wasClean) {
			state.open = false;
			logError('Socket closed');
			if (backoff < 30000) {
				backoff *= 2;
			}
		  if (!state.reconnecting) {
				logError('Scheduling reopen');
				state.reconnecting = true;
				setTimeout(setupWs, backoff);
			}
		};
    ws.onopen = function() {
			state.open = true;
		  logInfo("Socket opened");
		  backoff = 500;
			if (state.started) {
				for (var url in subscriptions) {
					logDebug('Re-subscribing to', url);
					ws.sendIfReady(JSON.stringify({
						Type: 'Subscribe',
						Object: {
							URI: url,
						},
					}));
				}
			} else {
				state.started = true;
				start();
			}
		};
		ws.onerror = function(err) {
			state.open = false;
		  logError('WebSocket error', err);
		  if (backoff < 30000) {
				backoff *= 2;
			}
			if (!state.started) {
			  state.started = true;
				start();
			}
			if (!state.reconnecting) {
				logError('Scheduling reopen');
				state.reconnecting = true;
				setTimeout(setupWs, backoff);
			}
		};
		ws.onmessage = function(ev) {
			var mobj = JSON.parse(ev.data);
			if (mobj.Type == 'RPC') {
			  var rpcCall = rpcCalls[mobj.Method.Id];
				if (rpcCall != null) {
				  rpcCall(mobj.Method.Data);
				}
			} else {
				if (mobj.Object.URI != null) {
					var subscription = subscriptions[mobj.Object.URI];
					if (subscription != null) {
						logDebug('Got', mobj.Type, mobj.Object.URI, 'from websocket');
						logTrace(mobj.Object.Data);
						if (mobj.Type == 'Delete') {
							if (subscription.model.models != null) {
								_.each(mobj.Object.Data, function(element) {
									var model = subscription.model.get(element.Id);
									subscription.model.remove(model, { silent: true });
								});
								subscription.model.trigger('reset');
							} else {
								logError("Don't know how to handle Deletes on backbone.Models");
							}
						} else {
							if (subscription.options != null && subscription.options.success != null) {
								subscription.options.success(mobj.Object.Data, null, subscription.options);
								delete(subscription.options.success);
							} else {
								subscription.model.set(mobj.Object.Data, { remove: mobj.Type == 'Fetch', reset: true });
							}
						}
						if (_.result(subscription.model, 'localStorage')) {
							localStorage.setItem(mobj.Object.URI, JSON.stringify(subscription.model));
							logDebug('Stored', mobj.Object.URI, 'in localStorage');
						}
					} else {
						logError("Received", mobj, "for unsubscribed URI", mobj.Object.URI);
					}
				}
			}
		};
	};
	setupWs();

	Backbone.Collection.prototype.close = function() {
		closeSubscription(this);
	};
	Backbone.Model.prototype.close = function() {
	  closeSubscription(this);
	};
	Backbone.Model.prototype.idAttribute = "Id";

	Backbone.sync = function(method, model, options) {
		var urlError = function() {
			throw new Error('A "url" property or function must be specified');
		};
		var urlBefore = options.url || _.result(model, 'url') || urlError(); 
		if (method == 'read') {
			var cached = localStorage.getItem(urlBefore);
			if (cached != null) {
				logDebug('Loaded', urlBefore, 'from localStorage');
				var data = JSON.parse(cached);
				logTrace(data);
				var success = options.success;
				options.success = null;
				success(data, null, options);
			}
		}
		if (method == 'read') {
			logDebug('Subscribing to', urlBefore);
			subscriptions[urlBefore] = {
			  model: model,
				options: options,
			};
			ws.sendIfReady(JSON.stringify({
				Type: 'Subscribe',
				Object: {
					URI: urlBefore,
				},
			}));
		} else if (method == 'create') {
		  logDebug('Creating', urlBefore);
			ws.sendIfReady(JSON.stringify({
			  Type: 'Create',
				Object: {
				  URI: urlBefore,
					Data: model,
				},
			}));
			if (options.success) {
			  var success = options.success;
				options.success = null;
			  success(model.toJSON(), null, options);
			}
		} else if (method == 'update') {
      logDebug('Updating', urlBefore);
			ws.sendIfReady(JSON.stringify({
			  Type: 'Update',
				Object: {
				  URI: urlBefore,
					Data: model,
				},
			}));
			if (options.success) {
			  var success = options.success;
				options.success = null;
			  success(model.toJSON(), null, options);
			}
		} else if (method == 'delete') {
		  logDebug('Deleting', urlBefore);
			ws.sendIfReady(JSON.stringify({
			  Type: 'Delete',
				Object: {
          URI: urlBefore,
				},
			}));
		} else {
			logError("Don't know how to handle " + method);
			if (options.error) {
			  options.error(model, "Don't know how to handle " + method, options);
			}
		}
	};
};


