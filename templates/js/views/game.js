window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

  render: function() {
		this.$el.html(this.template({ }));
		panZoom('.map');
		return this;
	},

});
