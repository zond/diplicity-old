
window.session = {};

window.session.online = false;
window.session.updateOnlineTag = function() {
	if (window.session.online) {
		$('.offline-tag').hide();
	} else {
		$('.offline-tag').show();
	}
};
window.session.setOnline = function(online) {
  window.session.online = online;
	window.session.updateOnlineTag();
};


Backbone.Model.prototype.idAttribute = "Id";

$(window).load(function() {

  $(window).on('hide.bs.collapse', function(ev) {
		$('div[href=#' + $(ev.target).attr('id') + ']').find('.glyphicon-chevron-down').removeClass('glyphicon-chevron-down').addClass('glyphicon-chevron-right');
	});

  $(window).on('show.bs.collapse', function(ev) {
		$('div[href=#' + $(ev.target).attr('id') + ']').find('.glyphicon-chevron-right').removeClass('glyphicon-chevron-right').addClass('glyphicon-chevron-down');
	});

	var AppRouter = Backbone.Router.extend({

		routes: {
			"": "currentGames",
			"forming": "formingGames",
			"finished": "finishedGames",
			"open": "openGames",
			"create": "createGame",
			"games/:id": "game",
			"games/:id/messages/:participants": "chat",
			"settings": "settings",
		},

		settings: function() {
			new SettingsView({
				el: $('#content'),
			}).doRender();
		},

		game: function(id) {
			new GameView({
				model: new GameState({
					Id: id,
				}),
				el: $('#content'),
			}).doRender();
		},

		chat: function(gameId, participants) {
		  new GameView({
			  model: new GameState({
				  Id: gameId,
				}),
				el: $('#content'),
				chatParticipants: participants,
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

		finishedGames: function() {
			new CurrentGameStatesView({ 
				el: $('#content'),
				filter_state: '{{.GameState "Started" }}',
				filter_label: "Finished",
			}).doRender();
		},

		formingGames: function() {
			new CurrentGameStatesView({ 
				el: $('#content'),
				filter_state: '{{.GameState "Created" }}',
				filter_label: "Forming",
			}).doRender();
		},

		openGames: function() {
			new OpenGameStatesView({ 
				el: $('#content'),
			}).doRender();
		},
	});


	var start = function(ev) {
		window.session.router = new AppRouter();

		window.session.user = new User();

		window.session.user.fetch();
		window.session.active_url = null;

		window.session.online = false;

		window.session.top_navigation = new TopNavigationView({
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
	  state_handler: function(state) {
		  window.session.setOnline(state.open);
		},
	  url: url, 
		start: start,
		token_producer: function(opts) {
		  $.ajax('/token', {
        success: function(data) {
				  opts.success(data.Encoded);
				},
				error: opts.error,
			});
		},
		cache_backend: jsCacheBackend(),
	});

});
