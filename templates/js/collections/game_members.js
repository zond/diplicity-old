window.GameMembers = Backbone.Collection.extend({

  url: '/games',

	model: GameMember,
	
	initialize: function() {
		var that = this;
		this.on('change', function() {
			$('ul.games').empty();
			that.each(function(member) {
			  $('.games').append('<li><a href="#">{0}</a></li>'.format(member.describe()));
			});
		});
	  
	},

});

