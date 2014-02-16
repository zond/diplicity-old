
window.session = {};

$(window).load(function() {

	var AppRouter = Backbone.Router.extend({

		routes: {
			"": "currentGames",
			"open": "openGames",
			"create": "createGame",
			"games/:id": "games",
			"settings": "settings",
		},

		settings: function() {
			new SettingsView({
				el: $('#content'),
			}).doRender();
		},

		games: function(id) {
			new GameView({
				model: new GameState({
					Id: id,
				}),
				el: $('#content'),
			}).doRender();
		},

		createGame: function() {
			new CreateGameView({ 
				el: $('#content'),
			}).doRender();
		},

		currentGames: function() {
			new CurrentGameStatesView({ 
				el: $('#content'),
			}).doRender();
		},

		openGames: function() {
			new OpenGameStatesView({ 
				el: $('#content'),
			}).doRender();
		},
	});

	window.session.router = new AppRouter();

	var start = function(ev) {
		window.session.user = new User();

		window.session.user.fetch();
		window.session.active_url = null;

		new TopNavigationView({
			el: $('#top-navigation'),
		}).doRender();

		window.session.bottom_navigation = new BottomNavigationView({
			el: $('#bottom-navigation'),
			buttons: mainButtons,
		}).doRender();
	
		Backbone.history.start({ 
			pushState: true,
		});

    navigate(Backbone.history.fragment || '/');
	};

	var match = /^.*:\/\/([^\/]*)\//.exec(window.location.href);
	var url = "ws://" + match[1] + "/ws";

	wsBackbone({
	  url: url, 
		start: start,
		cacheBackend: jsCacheBackend(),
	});

});
