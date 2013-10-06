
function dippyMap(container) {
	var that = this;
	var el = null;
	var SVG = "http://www.w3.org/2000/svg";
	if (container.find('svg').length > 0) {
	  el = container.find('svg')[0];
	}
	that.centerOf = function(province) {
		var center = $(el).find('#' + selEscape(province) + "Center").first();
		var match = /^M\s+([\d.]+),([\d.]+)\s+/.exec(center.attr('d'));
		var x = Number(match[1]);
		var y = Number(match[2]);
		var parentTransform = center.parent().attr("transform");
		if (parentTransform != null) {
			var transMatch = /^translate\(([\d.-]+),\s*([\d.-]+)\)$/.exec(parentTransform);
			x += Number(transMatch[1]) - 1.5;
			y += Number(transMatch[2]) - 2;
		}
		return new Poi(x,y);
	};
	that.showProvinces = function() {
		$(el).find('#provinces')[0].removeAttribute('style');
	};
	that.copySVG = function(sourceId) {
		var source = $('#' + sourceId + ' svg').first().clone();
		container[0].appendChild(source[0]);
		el = container.find('svg')[0];
	};
	that.colorProvince = function(province, color) {
		var path = $(el).find('#' + selEscape(province))[0];
		path.removeAttribute('style');
		path.setAttribute('fill', color);
		path.setAttribute('fill-opacity', '0.8');
	};
  that.hideProvince = function(province) {
		var path = $(el).find('#' + selEscape(province))[0];
		path.removeAttribute('style');
		path.setAttribute('fill', '#ffffff');
		path.setAttribute('fill-opacity', '0');
	};
	that.blinkProvince = function(province) {
		var prov = $(el).find('#' + selEscape(province)).first()[0];
		prov.setAttribute("stroke", 'red');
		prov.setAttribute("stroke-width", '8');
		return function() {
			prov.setAttribute("stroke", 'none');
		};
	};
	that.addClickListener = function(province, handler) {
		var prov = $(el).find('#' + selEscape(province)).first();
		var copy = prov.clone()[0];
		copy.setAttribute("id", prov.attr('id') + "_click");
		copy.setAttribute("style", "fill:#000000;fill-opacity:0;stroke:none;");
		copy.setAttribute("stroke", "none");
		copy.removeAttribute("transform");
		var x = 0;
		var y = 0;
		var curr = prov[0];
		while (curr != null && curr.getAttribute != null) {
			var trans = curr.getAttribute("transform");
			if (trans != null) {
				var transMatch = /^translate\(([\d.-]+),\s*([\d.-]+)\)$/.exec(trans);
				x += Number(transMatch[1]);
				y += Number(transMatch[2]);
			}
			curr = curr.parentNode;
		}
		copy.setAttribute("transform", "translate(" + x + "," + y + ")");
		el.appendChild(copy);
		var ham = $(copy).hammer({
		  prevent_default: true,
		});
		ham.bind('tap', handler);
		return function() {
			ham.unbind('tap', handler);
		};
	};
	that.addCross = function(province, color) {
		var boundF = 10;
		var loc = that.centerOf(province);
		var path = document.createElementNS(SVG, "path");
		path.setAttribute("style", "fill:none;stroke:" + color + ";stroke-width:5;stroke-miterlimit:4;stroke-opacity:0.9;");
		path.setAttribute("d", "M " + (loc.x - boundF) + "," + (loc.y - boundF) + " L " + (loc.x + boundF) + "," + (loc.y + boundF) + " M " + (loc.x - boundF) + "," + (loc.y + boundF) + " L " + (loc.x + boundF) + "," + (loc.y - boundF));
		el.appendChild(path);
	};
	that.addOrder = function(order, color) {
	  if (order[1] == 'Disband') {
		  addCross(order[0], color);
		}
	};
	that.addUnit = function(sourceId, province, color, dislodged, build) {
		var shadow = $('#' + sourceId).find('#shadow').first().clone();
		var hullQuery = $('#' + sourceId + ' svg').find('#hull');
		var bodyQuery = $('#' + sourceId + ' svg').find('#body');
		var loc = that.centerOf(province);
		var unit = null;
		var opacity = 1;
		if (dislodged) {
			loc.x += 5;
			loc.y += 5;
			opacity = 0.73;
		}
		loc.y -= 11;
		if (hullQuery.length > 0) {
			unit = hullQuery.first().clone();
			loc.x -= 27;
		} else {
			unit = bodyQuery.first().clone();
			loc.x -= 16;
		}
		shadow.attr("transform", "translate(" + loc.x + ", " + loc.y + ")");
		unit.attr("transform", "translate(" + loc.x + ", " + loc.y + ")");
		if (build) {
			color = '#000000';
		}
		unit.attr("style", "fill:" + color + ";fill-opacity:" + opacity + ";stroke:#000000;stroke-width:1;stroke-miterlimit:4;stroke-opacity:1;stroke-dasharray:none");
		el.appendChild(shadow[0]);
		el.appendChild(unit[0]);
	};
	return that;
}




