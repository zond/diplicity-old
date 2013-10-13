
window.session = {};

$(window).load(function() {
	var start = function(ev) {
		window.session.user = new User();

		var AppRouter = Backbone.Router.extend({

			routes: {
				"": "currentGames",
				"open": "openGames",
				"create": "createGame",
				"menu": "menu",
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

			menu: function() {
				new MenuView({ 
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

		window.session.user.fetch();
		window.session.active_url = null;

		new TopNavigationView({
		  el: $('#top_navigation'),
		}).doRender();
		window.session.bottom_navigation = new BottomNavigationView({
		  el: $('#bottom_navigation'),
			buttons: mainButtons,
		}).doRender();

		window.session.router = new AppRouter();
		Backbone.history.start({ 
			pushState: true,
		});

    navigate(Backbone.history.fragment || '/');
	};
	var match = /^.*:\/\/([^\/]*)\//.exec(window.location.href);
	var url = "ws://" + match[1] + "/ws";
	wsBackbone(url, start);

});
