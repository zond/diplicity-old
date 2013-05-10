
$(window).load(function() {
  $(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
	  if (jqXHR.status == 401) {
		  console.log(jqXHR.getResponseHeader("Location"));
		}
	});
  panZoom('.map');
	var user = new User();
	var gameMembers = new GameMembers();
	user.bind('change', function() {
	  if (user.get('email') != null && user.get('email') != '') {
			gameMembers.fetch();
		}
	});
	user.fetch();
});

