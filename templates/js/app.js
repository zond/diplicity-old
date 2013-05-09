
$(document).ready(function() {
  panZoom('.map');
	var games = new Games();
	games.fetch();
});

