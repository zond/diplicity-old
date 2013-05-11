window.GameMembers = Backbone.Collection.extend({

  url: '/games',

	model: GameMember,

	initialize: function() {
		var that = this;
    _.bindAll(this, 'render');
		this.bind("change", this.render);
		this.bind("reset", this.render);
		this.bind("add", this.render);
		this.bind("remove", this.render);
	},

	render: function() {
	  $('ul.games').empty();
		this.each(function(member) {
		  console.log('rendering', member, 'to', $('ul.games'));
			$('ul.games').append('<li><a href="#">{0}</a></li>'.format(member.describe()));
		});
		$('ul.games').listview('refresh');
	},

});

