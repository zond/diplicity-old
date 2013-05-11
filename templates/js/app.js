
$(window).load(function() {

	$(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
		if (jqXHR.status == 401) {
			console.log(jqXHR.getResponseHeader("Location"));
		}
	});
	panZoom('.map');

	var gameMembers = new GameMembers();
	var user = new User({}, {
		gameMembers: gameMembers,
	});
	user.fetch();

  $('.create-game-button').on('click', function(ev) {
	  var form = $(ev.target).closest('.create-game-form');
		gameMembers.create({
		  game: {
				variant: form.find('select.create-game-variant').val(),
				private: form.find('select.create-game-private').val() == 'true',
			},
		}, {
		  success: function() {
				$.mobile.changePage('#home');
			},
		});
	});

});

