window.Games = Backbone.Collection.extend({

  url: function() {
		return '/games';
	},

	model: Game,

});

