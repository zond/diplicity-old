window.CurrentGameMembersView = BaseView.extend({

  template: _.template($('#current_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		window.session.user.bind('change', this.doRender);
		window.session.currentGameMembers.bind("reset", this.doRender);
		window.session.currentGameMembers.bind("add", this.doRender);
		window.session.currentGameMembers.bind("remove", this.doRender);
	},

	onClose: function() {
		window.session.user.unbind('change', this.doRender);
	  window.session.currentGameMembers.unbind('reset', this.doRender);
	  window.session.currentGameMembers.unbind('add', this.doRender);
	  window.session.currentGameMembers.unbind('remove', this.doRender);
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
