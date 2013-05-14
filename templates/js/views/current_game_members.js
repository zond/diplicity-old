window.CurrentGameMembersView = Backbone.View.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'render', 'refetch');
		this.user = options.user;
		this.user.bind('change', this.refetch);
		this.collection.bind("reset", this.render);
		this.collection.bind("add", this.render);
		this.collection.bind("remove", this.render);
		this.children = [];
	},

	clean: function() {
	  _.each(this.children, function(child) {
		  child.onClose();
		});
		this.children = [];
	},

	refetch: function() {
		if (this.user.loggedIn()) {
		  this.collection.fetch();
		}
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
