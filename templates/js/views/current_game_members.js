window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'refetch');
		this.user = options.user;
		this.user.bind('change', this.refetch);
		this.collection.bind("reset", this.doRender);
		this.collection.bind("add", this.doRender);
		this.collection.bind("remove", this.doRender);
	},

	refetch: function() {
		if (this.user.loggedIn()) {
		  this.collection.fetch();
		}
	},

  render: function() {
	  console.log('re-rendering current game members');
	  var that = this;
		that.clean();
		that.$el.html(that.template({}));
		that.collection.forEach(function(model) {
			that.$el.append(new GameMemberView({ 
				model: model,
			}).doRender().el);
		});
		return that;
	},

});
