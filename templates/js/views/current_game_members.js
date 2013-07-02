window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		this.listenTo(window.session.currentGameMembers, "reset", this.doRender);
		this.listenTo(window.session.currentGameMembers, "add", this.doRender);
		this.listenTo(window.session.currentGameMembers, "remove", this.doRender);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({}));
		window.session.currentGameMembers.forEach(function(model) {
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
