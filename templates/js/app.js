
window.session = {};

$(window).load(function() {
  var match = /^.*:\/\/(.*)\//.exec(window.location.href);
	var url = "ws://" + match[1] + "/ws";
	var start = function(ev) {
		window.session.user = new User();

		var AppRouter = Backbone.Router.extend({

			routes: {
				"": "currentGames",
				"open": "openGames",
				"create": "createGame",
				"menu": "menu",
				"game": "game",
			},

			game: function() {
				new GameView({
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

		window.session.router = new AppRouter();
		Backbone.history.start({ 
			pushState: true,
		});
		window.session.user.fetch();

		new TopNavigationView({
		  el: $('#top_navigation'),
		}).doRender();
		window.session.bottom_navigation = new BottomNavigationView({
		  el: $('#bottom_navigation'),
		}).doRender();

    navigate(Backbone.history.fragment || '/');
	};
  wsBackbone(url, start);

});
