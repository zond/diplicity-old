
$(window).load(function() {

	$(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
		if (jqXHR.status == 401) {
			console.log(jqXHR.getResponseHeader("Location"));
		}
	});
	panZoom('.map');

	var currentGameMembers = new GameMembers([], { url: '/games/member' });
	var user = new User();

  new CurrentGameMembersView({ 
	  el: $('.games'),
	  collection: currentGameMembers,
		user: user,
	}).render();

	new CreateGameView({
    el: $('.create-game'),
	  collection: currentGameMembers,
	}).render();

	new OpenGameMembersView({
	  el: $('.join-game'),
	  user: user,
	  currentGameMembers: currentGameMembers,
	}).render();

	user.fetch();

});

