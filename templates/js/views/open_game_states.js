window.OpenGameStatesView = BaseView.extend({

  template: _.template($('#open_game_states_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/open' });
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
		that.collection.forEach(function(model) {
		  var memberView = new GameStateView({ 
				model: model,
				editable: false,
				button_text: '{{.I "Join" }}',
				button_action: function() {
				  console.log('going to save',model,'with id',model.id);
					model.save(null, {
						success: function() {
							that.collection.remove(model);
							window.session.router.navigate('', { trigger: true });
						},
					});
				},
			}).doRender();
			memberView.$el.attr('data-role', 'collapsible');
			that.$el.append(memberView.el);
		});
		that.$el.trigger('pagecreate');
		return that;
	},

});
