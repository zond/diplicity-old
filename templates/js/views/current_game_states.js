window.CurrentGameStatesView = BaseView.extend({

  template: _.template($('#current_game_states_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/current' });
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
		this.fetch(this.collection);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		this.collection.forEach(function(model) {
		  var memberView = new GameStateView({ 
				model: model,
				editable: false,
				button_text: '{{.I "Leave" }}',
				button_action: function() {
					model.destroy();
				},
			}).doRender();
			that.$el.append(memberView.el);
		});
		return that;
	},

});
