window.OthersGameStatesView = BaseView.extend({

  template: _.template($('#others_game_states_underscore').html()),

	initialize: function(options) {
		this.path = options.path;
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/' + options.path });
		this.listenTo(this.collection, "sync", this.doRender);
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
		this.fetch(this.collection);
	},

  render: function() {
	  var that = this;
		var nav = [
			[
				{
					url: '/open',
					label: '{{.I "Open" }}',
					activate: function() {
						return that.path == 'open';
					},
				},
				{
					url: '/closed',
					label: '{{.I "Closed" }}',
					activate: function() {
						return that.path == 'closed';
					},
				},
				{
					url: '/finished',
					label: '{{.I "Finished" }}',
					activate: function() {
						return that.path == 'finished';
					},
				},
			],
			mainButtons[0],
		];
		navLinks(nav);
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		that.collection.forEach(function(model) {
			var save_call = function() {
				model.save(null, {
					success: function() {
						navigate('/');
					},
				});
			};
			var stateView = new GameStateView({ 
				model: model,
				parentId: 'others-games',
				editable: false,
			}).doRender();
			that.$('#others-games').append(stateView.el);
		});
		if (window.session.user.loggedIn() && that.collection.length == 0 && that.path == 'open') {
		  that.$el.append('<a href="/create" class="btn btn-primary btn-lg btn-block">{{.I "No open games, why not create one?" }}</a>');
		}
		that.$('#others-games').css('margin-bottom', $('#bottom-navigation').height() + 'px');
		return that;
	},

});
