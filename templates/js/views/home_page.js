window.HomePageView = Backbone.View.extend({

  template: _.template($('#home_page_underscore').html()),


	initialize: function(options) {
	this.user=options.user;
	this.currentGameMembers=options.currentGameMembers;
	},

  render: function() {
  
		this.$el.html(this.template({
			user:this.user,
			
		}));
		new CurrentGameMembersView({ 
			el: this.$('.homePageGames'),
			collection: this.collection,
			user: this.user,
		}).render();
		this.$el.trigger('create');
		this.delegateEvents();
		return this;
	},

});
