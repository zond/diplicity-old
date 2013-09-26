window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(this.model, 'change', this.doRender);
		this.fetch(this.model);
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({ 
		}));
		console.log(this.model.attributes);
		if (this.model.get('Phase') != null) {
			var state_view = new GameStateView({ 
				parentId: 'current_game',
				editable: false,
				model: that.model,
			}).doRender();
			that.$('#current_game').append(state_view.el);
			that.$('.map-container').show();
		} else {
			that.$('.map-container').hide();
		}
		panZoom('.map');
		return that;
	},

});
