window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'refetch');
		this.user = options.user;
		this.user.bind('change', this.refetch);
		this.collection.bind("reset", this.doRender);
		this.collection.bind("add", this.doRender);
		this.collection.bind("remove", this.doRender);
		this.refetch();
	},

	onClose: function() {
		this.user.unbind('change', this.refetch);
	  this.collection.unbind('reset', this.doRender);
	  this.collection.unbind('add', this.doRender);
	  this.collection.unbind('remove', this.doRender);
	},

	refetch: function() {
		if (this.user.loggedIn()) {
		  this.collection.fetch();
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({}));
		that.collection.forEach(function(model) {
		  var memberView = new GameMemberView({ 
				model: model,
			}).doRender();
			memberView.$el.attr('data-role', 'collapsible');
			that.$el.append(memberView.el);
		});
		return that;
	},

});
