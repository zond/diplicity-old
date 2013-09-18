
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

function panZoom(selector) {
	var MAX_ZOOM = 4;
  $(selector).each(function(ind, el) {
	  var element = $(el);
		var scale = 1;
		var zoom = 1;
		var deltaX = 0;
		var deltaY = 0;
		var dragX = 0;
		var dragY = 0;
		var transforming = false;
		var container = element.parent().hammer({ 
			prevent_default: true,
		});
		var state = function() {
		  return JSON.stringify({
			  scale: scale,
				zoom: zoom,
				deltaX: deltaX,
				deltaY: deltaY,
				dragX: dragX,
				dragY: dragY,
			}, null, 2);
		};
		var execute = function() {
			element.css('-webkit-transform', 'translate3d(' + parseInt(deltaX + dragX) + 'px,' + parseInt(deltaY + dragY) + 'px,0px) scale3d(' + (scale * zoom) + ',' + (scale * zoom) + ',1)');
		};
		container.bind('mousewheel', function(e) {
		  e.preventDefault();
			var wantedZoom = 1;
			if (e.originalEvent.wheelDelta > 0) {
			  wantedZoom = 1 + e.originalEvent.wheelDelta / $(window).height();
				if (wantedZoom > 2) {
				  wantedZoom = 2;
				}
			} else {
			  wantedZoom = 1 - e.originalEvent.wheelDelta / -$(window).height();
				if (wantedZoom < 0.5) {
				  wantedZoom = 0.5;
				}
			}
			if ((wantedZoom > 1 && scale * wantedZoom < MAX_ZOOM) || (wantedZoom < 1 && scale * wantedZoom > (1 / MAX_ZOOM))) {
				scale = scale * wantedZoom;
				execute();
			}
		});
		container.bind('drag', function(e) {
		  if (!transforming) {
				dragX = e.gesture.deltaX;
				dragY = e.gesture.deltaY;
				var bottom = deltaY + dragY + element.height() * scale * zoom + element.height() * (1 - scale * zoom) * 0.5;
				if (bottom < $(window).height() / 2) {
				  deltaY = $(window).height() / 2 + deltaY - bottom;
				}
				var top = deltaY + dragY + element.height() * (1 - scale * zoom) * 0.5;
				if (top > $(window).height() / 2) {
				  deltaY = $(window).height() / 2 + deltaY - top;
				}
				var left = deltaX + dragX + element.width() * (1 - scale * zoom) * 0.5;
				if (left > $(window).width() / 2) {
				  deltaX = $(window).width() / 2 + deltaX - left;
				}
				var right = deltaX + dragX + element.width() * scale * zoom + element.width() * (1 - scale * zoom) * 0.5;
				if (right < $(window).width() / 2) {
				  deltaX = $(window).width() / 2 + deltaX - right;
				}
				execute();
			}
		});
		container.bind('dragend', function(e) {
		  if (!transforming) {
				deltaX += dragX;
				deltaY += dragY;
				dragX = 0;
				dragY = 0;
			}
		});
		container.bind('transformstart', function(e) {
		  transforming = true;
		});
		container.bind('transform', function(e){
		  if ((e.gesture.scale > 1 && scale * e.gesture.scale < MAX_ZOOM) || (e.gesture.scale < 1 && scale * e.gesture.scale > (1 / MAX_ZOOM))) {
				zoom = e.gesture.scale;
				execute();
			}
		});
		container.bind('transformend', function(e) {
		  scale = scale * zoom;
			zoom = 1;
			transforming = false;
		});
	});
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

function wsBackbone(ws) {
  var subscriptions = {};
	var closeSubscription = function(that) {
		var url = _.result(that, 'url') || urlError(); 
		if (subscriptions[url] != null) {
			logDebug('Unsubscribing from', url);
			ws.send(JSON.stringify({
			  Type: 'Unsubscribe',
				Subscribe: {
				  URI: url,
				},
			}));
			delete(subscriptions[url]);
		}
	};

	Backbone.Collection.prototype.close = function() {
	  closeSubscription(this);
	};
	Backbone.Model.prototype.close = function() {
	  closeSubscription(this);
	};
	Backbone.Model.prototype.idAttribute = "Id";

	var oldBackboneSync = Backbone.sync;
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
			ws.send(JSON.stringify({
				Type: 'Subscribe',
				Object: {
					URI: urlBefore,
				},
			}));
		} else if (method == 'create') {
		  logDebug('Creating', urlBefore);
			ws.send(JSON.stringify({
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
			ws.send(JSON.stringify({
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
			ws.send(JSON.stringify({
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
	var oldOnmessage = ws.onmessage;
	ws.onmessage = function(ev) {
	  var mobj = JSON.parse(ev.data);
		if (mobj.Object.URI != null) {
			var subscription = subscriptions[mobj.Object.URI];
			if (subscription != null) {
				logDebug('Got', mobj.Object.URI, 'from websocket');
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
		if (oldOnmessage != null) {
		  oldOnmessage(ev);
		}
	};
};

function variants() {
	var rval = [];
	{{range .Variants}}rval.push({
		id: '{{.Id}}',
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
};

function variantName(id) {
	{{range .Variants}}if (id == '{{.Id}}') {
		return '{{.Translation}}';
	}
	{{end}}
	return null;
};

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

defaultVariant = 'standard';
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
				window.session.router.navigate($(el).attr('href'), { trigger: true });
			});
		});	
	},

	createJQM: function() {
		if (this.$el.attr('data-role') == 'page') {
			this.$el.trigger('pagecreate');
		} else {
			this.$el.trigger('create');
		}
	},

	doRender: function() {
		this.cleanChildren();
		if (this.chain.length > 0) {
			this.chain[this.chain.length - 1].addChild(this);
		} else if (this.el != null) {
		  if (this.el.CurrentBaseView != null) {
			  if (this.el.CurrentBaseView.cid == this.cid) {
				  this.cleanChildren();
				} else {
					this.el.CurrentBaseView.clean();
				}
			}
			this.el.CurrentBaseView = this;
		}
		this.chain.push(this);
		this.render();
		this.chain.pop();
		this.fixNavigateLinks();
		this.createJQM();
		return this;
	},

	clean: function() {
		if (typeof(this.onClose) == 'function') {
			this.onClose();
		}
		this.cleanChildren();
		this.stopSubscribing();
		this.stopListening();
	},

	stopSubscribing: function() {
		if (this.subscriptions != null) {
			_.each(this.subscriptions, function(subs) {
				subs.close();
			});
		}
		this.children = [];
	},

	cleanChildren: function() {
		if (this.children != null) {
			_.each(this.children, function(child) {
				child.clean();
			});
		}
		this.children = [];
	},

});


