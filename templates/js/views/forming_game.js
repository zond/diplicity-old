window.FormingGameView = Backbone.View.extend({

  template: _.template($('#forming_game_underscore').html()),

	initialize: function() {
	  _.bindAll(this, 'render');
		this.model.bind('change', this.render);
	},

  render: function() {
    this.$el.html(this.template({
		  model: this.model,
		}));
		return this;
	},

});
