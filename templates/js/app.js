
window.applicationCache.addEventListener('downloading', function(event) {
  if (window.session != null && window.session.bottom_navigation != null) {
	  window.session.bottom_navigation.showPercent(0);
	}
}, false);


window.applicationCache.addEventListener('progress', function(event) {
  if (window.session != null && window.session.bottom_navigation != null) {
	  window.session.bottom_navigation.showPercent(parseInt(100.0 * (parseFloat(event.loaded) / parseFloat(event.total))));
	}
}, false);

window.applicationCache.addEventListener('updateready', function(event) {
	window.localStorage.clear();
  if (window.session != null && window.session.bottom_navigation != null) {
	  window.session.bottom_navigation.showPercent(100);
	}
}, false);

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
			"": "myRunning",
			"mine/forming": "myForming",
			"mine/finished": "myFinished",
			"open": "open",
			"closed": "closed",
			"finished": "finished",
			"create": "createGame",
			"games/:id": "game",
			"games/:id/messages/:participants": "chat",
			"settings": "settings",
			"map/:variant": "map",
		},

		map: function(variant) {
			new MapView({
				variant: variant,
				el: $('#view'),
			}).doRender();
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

		myRunning: function() {
			new MyGameStatesView({ 
				el: $('#content'),
				filter_state: {{.GameState "Started" }},
			}).doRender();
		},

		myFinished: function() {
			new MyGameStatesView({ 
				el: $('#content'),
				filter_state: {{.GameState "Ended" }},
			}).doRender();
		},

		myForming: function() {
			new MyGameStatesView({ 
				el: $('#content'),
				filter_state: {{.GameState "Created" }},
			}).doRender();
		},

		open: function() {
			new OthersGameStatesView({ 
				el: $('#content'),
				path: 'open',
			}).doRender();
		},

		closed: function() {
			new OthersGameStatesView({ 
				el: $('#content'),
				path: 'closed',
			}).doRender();
		},

		finished: function() {
			new OthersGameStatesView({ 
				el: $('#content'),
				path: 'finished',
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

	var match = /^(.*):\/\/([^\/]*)\//.exec(window.location.href);
  var prefix = "ws";
	if (match[1] == "https") {
    prefix = "wss";
	}
	var url = prefix + "://" + match[2] + "/ws";

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
