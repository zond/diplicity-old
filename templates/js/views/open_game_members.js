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
		this.children = [];
	},

	refetch: function() {
		if (this.user.loggedIn()) {
		  this.collection.fetch();
		}
	},

	clean: function() {
	  _.each(this.children, function(child) {
		  child.onClose();
		});
		this.children = [];
	},

  render: function() {
	  var that = this;
	  that.clean();
		that.$el.html(that.template({}));
		that.collection.forEach(function(model) {
			that.$el.append(new GameMemberView({ 
				model: model,
				parent: that,
			}).render().el);
		});
		that.$el.trigger('create');
		return that;
	},

});
