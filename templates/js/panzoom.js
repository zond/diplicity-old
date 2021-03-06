function panZoom(selector) {
	var MAX_ZOOM = 4;
  $(selector).each(function(ind, el) {
	  if (el._isPanZoomed) {
		  return
		}
		el._isPanZoomed = true;
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
		var viewToMap = function(p) {
		  return p.sub(new Poi(parseInt(deltaX + dragX), parseInt(deltaY + dragY))).div(scale * zoom);
		};
		var mapToView = function(p) {
		  return p.mul(scale * zoom).add(new Poi(parseInt(deltaX + dragX), parseInt(deltaY + dragY)));
		};
		var getMapCenter = function() {
		  return viewToMap(new Poi($(window).width() / 2, $(window).height() / 2));
		};
		var execute = function() {
			element.css('-webkit-transform', 'translate3d(' + parseInt(deltaX + dragX) + 'px,' + parseInt(deltaY + dragY) + 'px,0px) scale3d(' + (scale * zoom) + ',' + (scale * zoom) + ',1)');
			element.css('-webkit-transform-origin', '0 0');
		};
		deltaX = $(window).width() / 2 - element.width() / 2;
		deltaY = $(window).height() / 2 - element.height() / 2;
		execute();
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
				var oldCenter = getMapCenter();
				scale = scale * wantedZoom;
				var d = getMapCenter().sub(oldCenter).mul(scale * zoom);
				deltaX += d.x;
				deltaY += d.y;
				execute();
			}
		});
		container.bind('drag', function(e) {
		  if (!transforming) {
				dragX = e.gesture.deltaX;
				dragY = e.gesture.deltaY;
				var topLeft = mapToView(new Poi(0, 0));
				var bottomRight = mapToView(new Poi(element.width(), element.height()));
				if (bottomRight.y < $(window).height() / 2) {
				  deltaY = $(window).height() / 2 + deltaY - bottomRight.y;
				}
				if (topLeft.y > $(window).height() / 2) {
				  deltaY = $(window).height() / 2 + deltaY - topLeft.y;
				}
				if (topLeft.x > $(window).width() / 2) {
				  deltaX = $(window).width() / 2 + deltaX - topLeft.x;
				}
				if (bottomRight.x < $(window).width() / 2) {
				  deltaX = $(window).width() / 2 + deltaX - bottomRight.x;
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
				var oldCenter = getMapCenter();
				zoom = e.gesture.scale;
				var d = getMapCenter().sub(oldCenter).mul(scale * zoom);
				deltaX += d.x;
				deltaY += d.y;
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


