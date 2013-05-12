
$(window).load(function() {

	$(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
		if (jqXHR.status == 401) {
			console.log(jqXHR.getResponseHeader("Location"));
		}
	});
	panZoom('.map');

	var gameMembers = new GameMembers();
	var user = new User();

  new GameMembersView({ 
	  el: $('.games'),
	  collection: gameMembers,
		user: user,
	}).render();

	$('.create-game').append(new CreateGameView({
	  collection: gameMembers,
	}).render().el);

	$('.join-game').append(new JoinGameView({
	  user: user,
	  collection: gameMembers,
	}).render().el);

	user.fetch();

});

