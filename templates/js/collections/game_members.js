window.GameMembers = Backbone.Collection.extend({

	model: GameMember,

	initialize: function(data, options) {
	  this.url = options.url;
	},

});

