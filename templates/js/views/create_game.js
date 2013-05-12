window.CreateGameView = Backbone.View.extend({

  template: _.template($('#create_game_underscore').html()),

  events: {
	  "click .create-game-button": "createGame",
	},

	createGame: function(ev) {
		var that = this;
	  this.collection.create({
		  game: {
				variant: this.$('select.create-game-variant').val(),
				private: this.$('select.create-game-private').val() == 'true',
			}
		});
		$.mobile.changePage('#home');
	},

	initialize: function(options) {
	  _.bindAll(this, 'render');
	},

  render: function() {
		this.$el.html(this.template({}));
    this.$('.create-game-variant').selectmenu();
    this.$('.create-game-private').slider();
		this.$('.create-game-button').button();
		this.delegateEvents();
		return this;
	},

});
