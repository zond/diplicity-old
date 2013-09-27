
var oldConsoleLog = console.log
console.log = function() {
	var ary = Array.prototype.slice.call(arguments, 0); 
	ary.unshift(new Date());
	oldConsoleLog.apply(this, ary)
}

var logLevel = {{.LogLevel}};

function logFatal() {
  if (logLevel >= 0) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("FATAL");
	  console.log.apply(console, ary)
	}
}

function logError() {
  if (logLevel >= 1) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("ERROR");
	  console.log.apply(console, ary)
	}
}

function logInfo() {
  if (logLevel >= 2) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("INFO");
	  console.log.apply(console, ary)
	}
}

function logDebug() {
  if (logLevel >= 3) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("DEBUG");
	  console.log.apply(console, ary)
	}
}

function logTrace() {
  if (logLevel >= 4) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("TRACE");
	  console.log.apply(console, ary)
	}
}

String.prototype.format = function() {
	var args = arguments;
	return this.replace(/{(\d+)}/g, function(match, number) { 
		return typeof args[number] != 'undefined'
		? args[number]
		: match
		;
	});
};

function wsBackbone(url, start) {
	var ws = null;
  
	var subscriptions = {};
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

function allocationMethods() {
	var rval = [];
	{{range .AllocationMethods}}rval.push({
		id: '{{.Id}}',
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
}

function variants() {
	var rval = [];
	{{range .Variants}}rval.push({
		id: '{{.Id}}',
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
};

function allocationMethodName(id) {
 {{range .AllocationMethods}}if (id == '{{.Id}}') {
   return '{{.Translation}}';
 }
 {{end}}
 return null;
}

function variantName(id) {
	{{range .Variants}}if (id == '{{.Id}}') {
		return '{{.Translation}}';
	}
	{{end}}
	return null;
};

function variantNations(id) {
  {{range .Variants}}if (id == '{{.Id}}') {
    return {{.JSONNations}};
	}
	{{end}}
	return null;
}

function phaseTypes(variant) {
	{{range .Variants}}if (variant == '{{.Id}}') {
		var rval = [];
		{{range .PhaseTypes}}rval.push('{{.}}');
		{{end}}
		return rval;
	}
	{{end}}
	return [];
};

function chatFlagOptions() {
	var rval = [];
	{{range .ChatFlagOptions}}rval.push({
		id: {{.Id}},
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
};

defaultAllocationMethod = '{{.DefaultAllocationMethod}}';
defaultVariant = '{{.DefaultVariant}}';
defaultDeadline = 1440;
defaultChatFlags = {{.ChatFlag "White"}} | {{.ChatFlag "Conference"}} | {{.ChatFlag "Private"}};

deadlineOptions = [
	{ value: 5, name: '{{.I "5 minutes" }}' },
	{ value: 10, name: '{{.I "10 minutes" }}' },
	{ value: 20, name: '{{.I "20 minutes" }}' },
	{ value: 30, name: '{{.I "30 minutes" }}' },
	{ value: 60, name: '{{.I "1 hour" }}' },
	{ value: 120, name: '{{.I "2 hours" }}' },
	{ value: 240, name: '{{.I "4 hours" }}' },
	{ value: 480, name: '{{.I "8 hours" }}' },
	{ value: 720, name: '{{.I "12 hours" }}' },
	{ value: 1440, name: '{{.I "24 hours" }}' },
	{ value: 2880, name: '{{.I "2 days" }}' },
	{ value: 4320, name: '{{.I "3 days" }}' },
	{ value: 5760, name: '{{.I "4 days" }}' },
	{ value: 7200, name: '{{.I "5 days" }}' },
	{ value: 10080, name: '{{.I "1 week" }}' },
	{ value: 14400, name: '{{.I "10 days" }}' },
	{ value: 20160, name: '{{.I "2 weeks" }}' },
];

function deadlineName(value) {
  var found = _.find(deadlineOptions, function(opt) {
	  return opt.value == value;
	});
	if (found != null) {
	  return found.name;
	} else {
	  return '';
	}
};

var mainButtons = [
	{
	  url: '/',
		label: '{{.I "Games" }}',
	},
	{
	  url: '/open',
		label: '{{.I "Join" }}',
	},
	{
	  url: '/create',
		label: '{{.I "Create" }}',
	},
];

function navLinks(buttons) {
  window.session.bottom_navigation.navLinks(buttons);
};

function navigate(to) {
	window.session.active_url = to;
	window.session.router.navigate(to, { trigger: true });
	window.session.bottom_navigation.update();
	$('body').css({'margin-top': (($('.navbar-fixed-top').height()) + 1 )+'px'});
	$('body').css({'margin-bottom': (($('.navbar-fixed-bottom').height()) + 1 )+'px'});
}

window.BaseView = Backbone.View.extend({
 
	chain: [],

	fetch: function(obj) {
	  if (this.subscriptions == null) {
		  this.subscriptions = [];
		}
		this.subscriptions.push(obj);
		obj.fetch();
	},

	addChild: function(child) {
		if (this.children == null) {
			this.children = [];
		}
		this.children.push(child);
	},

	fixNavigateLinks: function() {
		this.$('a.navigate').each(function(ind, el) {
			$(el).bind('click', function(ev) {
				ev.preventDefault();
				navigate($(el).attr('href'));
			});
		});	
	},

	renderWithin: function(f) {
	  if (this.chain.length > 0 && this.chain[this.chain.length - 1].cid == this.cid) {
		  f();
		} else {
			this.chain.push(this);
			f();
			this.chain.pop();
		}
	},

	doRender: function() {
	  var that = this;
		that.cleanChildren();
		if (that.chain.length > 0) {
			that.chain[that.chain.length - 1].addChild(that);
		} else if (that.el != null) {
		  if (that.el.CurrentBaseView != null) {
			  if (that.el.CurrentBaseView.cid == that.cid) {
				  that.cleanChildren();
				} else {
					that.el.CurrentBaseView.clean();
				}
			}
			that.el.CurrentBaseView = that;
		}
		that.renderWithin(function() {
			that.render();
		});
		that.fixNavigateLinks();
		if (that.rendered) {
		  that.delegateEvents();
		}
	  that.rendered = true;
		return that;
	},

	clean: function(remove) {
		if (typeof(this.onClose) == 'function') {
			this.onClose();
		}
		this.cleanChildren(remove);
		this.stopSubscribing();
		if (remove) {
		  this.remove();
		} else {
			this.stopListening();
		}
	},

	stopSubscribing: function() {
		if (this.subscriptions != null) {
			_.each(this.subscriptions, function(subs) {
				subs.close();
			});
		}
		this.children = [];
	},

	cleanChildren: function(remove) {
		if (this.children != null) {
			_.each(this.children, function(child) {
				child.clean(remove);
			});
		}
		this.children = [];
	},

});


