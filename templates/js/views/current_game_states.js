window.CurrentGameStatesView = BaseView.extend({

  template: _.template($('#current_game_states_underscore').html()),

	initialize: function(options) {
	  this.filter_state = options.filter_state;
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/current' });
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
					url: '/',
					label: '{{.I "Running" }}',
					click: function(ev) {
					  ev.preventDefault();
					  navigate('/', true);
						that.filter_state = {{.GameState "Started"}};
						that.doRender();
					},
					activate: function() {
						return that.filter_state == {{.GameState "Started"}};
					},
				},
				{
					url: '/forming',
					label: '{{.I "Forming" }}',
					click: function(ev) {
					  ev.preventDefault();
					  navigate('/forming', true);
						that.filter_state = {{.GameState "Created"}};
						that.doRender();
					},
					activate: function() {
						return that.filter_state == {{.GameState "Created"}};
					},
				},
				{
					url: '/finished',
					label: '{{.I "Finished" }}',
					click: function(ev) {
					  ev.preventDefault();
					  navigate('/finished', true);
						that.filter_state = {{.GameState "Ended"}};
						that.doRender();
					},
					activate: function() {
						return that.filter_state == {{.GameState "Ended"}};
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
		  if (model.get('State') == that.filter_state) {
				that.$('#current-games').append(new GameStateView({ 
					model: model,
					parentId: "current-games",
					editable: false,
				}).doRender().el);
			}
		});
		if (window.session.user.loggedIn() && that.collection.length == 0) {
			that.$el.append('<a href="/open" class="btn btn-primary btn-lg btn-block">{{.I "Not member of any games right now, why not join one?" }}</a>');
		}
		return that;
	},

});
