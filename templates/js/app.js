
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
					el: $('#main'),
				}).doRender();
			},

			menu: function() {
				new MenuView({ el: $('#main') }).doRender();
			},

			createGame: function() {
				new CreateGameView({ 
					el: $('#main'),
				}).doRender();
			},

			currentGames: function() {
				new CurrentGameStatesView({ 
					el: $('#main'),
				}).doRender();
			},

			openGames: function() {
				new OpenGameStatesView({ 
					el: $('#main'),
				}).doRender();
			},
		});

		window.session.router = new AppRouter();
		Backbone.history.start({ 
			pushState: true,
		});
		window.session.user.fetch();

		window.session.router.navigate(Backbone.history.fragment || '');
	};
  wsBackbone(url, start);

});
