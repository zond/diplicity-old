
$(window).load(function() {
  $(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
	  if (jqXHR.status == 401) {
		  console.log(jqXHR.getResponseHeader("Location"));
		}
	});
  panZoom('.map');
	var user = new User();
	var userChanged = function() {
		if (user.get('email') == null || user.get('email') == '') {
			$('.login-button').css('display', 'block');
			$('.logout-button').css('display', 'none');
		} else {
			$('.login-button').css('display', 'none');
			$('.logout-button').css('display', 'block');
		}
	};
	user.on('sync', userChanged);
	user.fetch();
});

