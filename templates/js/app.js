
function log(s) {
  $('#log').html($('#log').html() + s + '<br/>');
}

function panZoom(id) {
  var el = document.getElementById(id);
	var t1s = null;
	var t2s = null;
	var t1 = null;
	var t2 = null;

	el.addEventListener('touchstart',function(event) {
	  _.each(event.changedTouches, function(ev) {
		  if (t1s == null) {
			  log('start 1');
			  t1s = ev;
			} else if (t2s == null && ev.identifier != t1s.identifier) {
			  log('start 2');
			  t2s = ev;
			}
		});
	}, false);

	el.addEventListener('touchmove',function(event) {
	  _.each(event.changedTouches, function(ev) {
		  if (t1s != null && ev.identitier == t1s.identifier) {
			  log('move 1');
			  t1 = ev;
			} else if (t2s != null && ev.identifier == t2s.identifier) {
			  log('move 2');
			  t2 = ev;
			}
		});
		if (t1s != null && t2s != null && t1 != null && t2 != null) {
			log('zoom!');
		  var before = Math.sqrt(Math.pow(t1s.pageX - t2s.pageX, 2) + Math.pow(t1s.pageY - t2s.pageY, 2));
			var after =  Math.sqrt(Math.pow(t1.pageX - t2.pageX, 2) + Math.pow(t1.pageY - t2.pageY, 2));
			$('#map').width($('#map').width() * (after / before));
		}
	}, false);

	el.addEventListener('touchend',function(event) {
	  _.each(event.changedTouches, function(ev) {
		  if (t1s != null && ev.identifier == t1s.identifier) {
			  log('stop 1');
			  t1s = null;
				t1 = null;
			} else if (t2s != null && ev.identifier == t2s.identifier) {
			  log('stop 2');
			  t2s = null;
				t2 = null;
			}
		});
	}, false);
}

$(document).ready(function() {
  panZoom('map');
});

