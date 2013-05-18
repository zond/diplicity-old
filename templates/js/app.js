
window.session = {};

$(window).load(function() {

	$(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
		if (jqXHR.status == 401) {
			console.log(jqXHR.getResponseHeader("Location"));
		}
	});
	
	var AppRouter = Backbone.Router.extend({

		routes: {
			"": "home",
			"open": "openGames",
			"create": "createGame",
			"menu": "menu",
			"game": "game",
		},

		game: function() {
		  new GameView({
			  el: $('#main'),
			}).doRender();
		},

    home: function() {
      new HomeView({ 
				el: $('#main'),
				user: user,
				collection: currentGameMembers,
			}).doRender();
		},

    menu: function() {
		  new MenuView({ el: $('#main') }).doRender();
		},

    createGame: function() {
		  new CreateGameView({ 
				el: $('#main'),
				collection: currentGameMembers,
			}).doRender();
		},

		openGames: function() {
			new OpenGameMembersView({ 
				el: $('#main'),
        user: user,
				currentGameMembers: currentGameMembers,
			}).doRender();
		},
	});

	var currentGameMembers = new GameMembers([], { url: '/games/member' });
	var user = new User();
	var router = new AppRouter();
	
	window.session.user = user;
	window.session.currentGameMembers = currentGameMembers;
	window.session.router = router;

	Backbone.history.start({ 
		pushState: true,
	});
	user.bind('sync', loginSync);
	user.fetch();

	router.navigate(Backbone.history.fragment || '');

});

