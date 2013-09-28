window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'provinceClicked');
		this.listenTo(this.model, 'change', this.doRender);
		this.fetch(this.model);
	},

	provinceClicked: function(prov) {
	  logInfo('clicked', prov);
	},

  render: function() {
		var that = this;
		navLinks([]);
		that.$el.html(that.template({ 
		}));
		if (this.model.get('Members') != null) {
			var state_view = new GameStateView({ 
				parentId: 'current_game',
				play_state: true,
				editable: false,
				model: that.model,
			}).doRender();
			that.$('#current_game').append(state_view.el);
		}
		that.model.render('.map', this.provinceClicked);
		panZoom('.map');
		return that;
	},

});
