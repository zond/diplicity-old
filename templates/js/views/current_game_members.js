window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameMembers([], { url: '/games/current' });
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
		this.fetch(this.collection);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({}));
		this.collection.forEach(function(model) {
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
