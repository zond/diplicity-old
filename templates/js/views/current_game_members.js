window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		window.session.user.bind('change', this.doRender);
		this.collection = new GameMembers([], { url: "/games/current" });
		this.collection.bind("reset", this.doRender);
		this.collection.bind("add", this.doRender);
		this.collection.bind("remove", this.doRender);
		this.collection.fetch();
	},

	onClose: function() {
		window.session.user.unbind('change', this.doRender);
	  this.collection.unbind('reset', this.doRender);
	  this.collection.unbind('add', this.doRender);
	  this.collection.unbind('remove', this.doRender);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({}));
		that.collection.forEach(function(model) {
		  var memberView = new GameMemberView({ 
				model: model,
				button_text: '{{.I "Leave" }}',
				button_action: function() {
					model.destroy();
				},
			}).doRender();
			memberView.$el.attr('data-role', 'collapsible');
			that.$el.append(memberView.el);
		});
		return that;
	},

});
