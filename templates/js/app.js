
$(document).ready(function() {
  $(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
	  if (jqXHR.status == 401) {
		  window.location.href = '{{.LoginURL}}';
		}
	});
  panZoom('.map');
	var user = new User();
	user.on('change', function() {
	  console.log("select which menu options to display depending on if we have a user");
	});
	user.fetch();
});

