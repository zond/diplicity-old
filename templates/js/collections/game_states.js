window.GameStates = Backbone.Collection.extend({

  localStorage: function() {
	  return this.url == '/games/current';
	},

	model: GameState,

	initialize: function(data, options) {
	  this.url = options.url;
	},

});

