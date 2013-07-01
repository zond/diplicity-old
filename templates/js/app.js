
window.session = {};

$(window).load(function() {

  var match = /^.*:\/\/(.*)\//.exec(window.location.href);
  var socket = new WebSocket("ws://" + match[1] + "/ws");
  wsBackbone(socket);
	socket.onopen = function(ev) {

		window.session.user = new User();
		window.session.currentGameMembers = new GameMembers([], { url: "/games/current" });

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

			openGames: function() {
				new OpenGameMembersView({ 
					el: $('#main'),
				}).doRender();
			},
		});

		window.session.router = new AppRouter();
		Backbone.history.start({ 
			pushState: true,
		});
		window.session.user.fetch();
		window.session.currentGameMembers.fetch();

		window.session.router.navigate(Backbone.history.fragment || '');
	};

});

