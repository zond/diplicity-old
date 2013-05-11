
$(window).load(function() {

	$(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
		if (jqXHR.status == 401) {
			console.log(jqXHR.getResponseHeader("Location"));
		}
	});
	panZoom('.map');

	var gameMembers = new GameMembers();
	var user = new User();

  $('.games').append(new GameMembersView({ 
	  collection: gameMembers,
		user: user,
	}).render().el);

	$('.create-game').append(new CreateGameView({
	  collection: gameMembers,
	}).render().el);

	user.fetch();

});

