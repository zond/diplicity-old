window.CurrentGameStatesView = BaseView.extend({

  template: _.template($('#current_game_states_underscore').html()),

	initialize: function(options) {
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
		navLinks(mainButtons);
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		that.collection.forEach(function(model) {
		  that.$('#current-games').append(new GameStateView({ 
				model: model,
				parentId: "current-games",
				editable: false,
			}).doRender().el);
		});
		if (window.session.user.loggedIn() && that.collection.length == 0) {
			that.$el.append('<a href="/open" class="btn btn-primary btn-lg btn-block">{{.I "Not member of any games right now, why not join one?" }}</a>');
		}
		return that;
	},

});
