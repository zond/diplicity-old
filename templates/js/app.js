
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

	new CreateGameView({
    el: $('.create-game'),
	  collection: gameMembers,
	}).render();
	$('.create-game').trigger('create');

	new JoinGameView({
	  el: $('.join-game'),
	  user: user,
	  collection: gameMembers,
	}).render();

	user.fetch();

});

