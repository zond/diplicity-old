window.GameMembers = Backbone.Collection.extend({

  localStorage: function() {
	  return this.url == '/games/current';
	},

	model: GameMember,

	initialize: function(data, options) {
	  this.url = options.url;
	},

});

