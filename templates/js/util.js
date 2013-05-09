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

var oldBackboneSync = Backbone.sync;
Backbone.sync = function(method, model, options) {
	if (method == 'read') {
		var urlBefore = options.url || _.result(model, 'url') || urlError(); 
    var cached = localStorage.getItem(urlBefore);
		if (cached != null) {
		  console.log('Fetched', urlBefore, 'from localStorage');
		  model.set(JSON.parse(cached));
		}
	} else if (!navigator.onLine) {
		throw new Error('Only reading models is allowed when offline');
	}
	if (navigator.onLine) {
		var oldSuccess = options.success;
		options.success = function(obj, stat, xhr) {
			var urlAfter = options.url || _.result(model, 'url') || urlError();
			localStorage.setItem(urlAfter, JSON.stringify(obj));
			console.log('Stored', urlAfter, 'in localStorage');
			oldSuccess(obj, stat, xhr);
		};
		oldBackboneSync(method, model, options);
	}
};

