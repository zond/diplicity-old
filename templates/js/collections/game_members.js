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
	  $('fieldset.games').empty();
		this.each(function(member) {
			$('fieldset.games').append(member.render());
		});
		$('fieldset.games').trigger('create');
	},

});

