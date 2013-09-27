function Poi(x, y) {
	this.x = x;
	this.y = y;
	this.add = function(p) {
		return new Poi(x + p.x, y + p.y);
	};  
	this.sub = function(p) {
		return new Poi(x - p.x, y - p.y);
	};  
	this.len = function() {
		return Math.sqrt(Math.pow(x, 2) + Math.pow(y, 2));
	};
	this.div = function(f) {
		return new Poi(x / f, y / f);
	};
	this.mul = function(f) {
		return new Poi(x * f, y * f);
	};
	this.orth = function() {
		return new Poi(-y, x);
	};
}

function Vec(p1, p2) {
	this.p1 = p1;
	this.p2 = p2;
	this.len = function() {
		return p2.sub(p1).len();
	};
	this.dir = function() {
		return p2.sub(p1).div(this.len());
	};
	this.orth = function() {
		return this.dir().orth();
	};
}

$.fn.centerOf = function(province) {
	var center = this.find('#' + province + "Center").first();
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

$.fn.copySVG = function(sourceId) {
  var source = $('#' + sourceId + ' svg').first().clone();
  $(this).each(function() {
	  this.appendChild(source[0]);
	});
}

$.fn.colorProvince = function(province, color) {
  this.each(function() {
	  var path = $(this).find('#' + selEscape(province))[0];
	  path.removeAttribute('style');
		path.setAttribute('fill', color);
		path.setAttribute('fill-opacity', '0.8');
	});
}

$.fn.hideProvince = function(province) {
  this.each(function() {
	  var path = $(this).find('#' + selEscape(province))[0];
		path.removeAttribute('style');
		path.setAttribute('fill', '#ffffff');
		path.setAttribute('fill-opacity', '0');
	});
}

$.fn.addUnit = function(sourceId, province, color, dislodged, build) {
  var shadow = $('#' + sourceId).find('#shadow').first().clone();
	var hullQuery = $('#' + sourceId + ' svg').find('#hull');
	var bodyQuery = $('#' + sourceId + ' svg').find('#body');
	var loc = this.centerOf(province.split('/')[0]);
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
  this.each(function() {
	  this.appendChild(shadow[0]);
		this.appendChild(unit[0]);
	});
};


