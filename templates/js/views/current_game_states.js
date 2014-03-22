window.CurrentGameStatesView = BaseView.extend({

  template: _.template($('#current_game_states_underscore').html()),

	events: {
	  'click .filter-button': 'changeFilter',
	},

	initialize: function(options) {
		this.filter_label = options.filter_label || 'Running';
		this.filter_state = options.filter_state || '{{.GameState "Started" }}';
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/current' });
		this.listenTo(this.collection, "sync", this.doRender);
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
		this.fetch(this.collection);
	},

	changeFilter: function(ev) {
	  ev.preventDefault();
		this.filter_state = $(ev.target).attr('data-state');
		this.filter_label = $(ev.target).text();
    navigate($(ev.target).attr('href'), true);
		this.doRender();
	},

  render: function() {
	  var that = this;
		navLinks(mainButtons);
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		that.$('.filter-button').removeClass('btn-primary');
		that.$('.filter-' + that.filter_label).addClass('btn-primary');
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
