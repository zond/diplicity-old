window.OpenGameMembersView = Backbone.View.extend({

  template: _.template($('#open_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'render', 'refetch');
		this.user = options.user;
		this.user.bind('change', this.refetch);
		this.gameMembers = options.gameMembers;
		this.collection = new GameMembers([], { url: '/games/open' });
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);
	},

	refetch: function() {
		if (this.user.loggedIn()) {
		  this.collection.fetch();
		}
	},

  render: function() {
		this.$el.html(this.template({}));
	  var that = this;
		this.collection.forEach(function(model) {
			that.$el.append(new GameMemberView({ model: model }).render().el);
		});
		this.$el.trigger('create');
		return this;
	},

});
