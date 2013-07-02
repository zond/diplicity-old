window.CreateGameView = BaseView.extend({

  template: _.template($('#create_game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		var deadlines = {};
		var chatFlags = {};
		_.each(phaseTypes(defaultVariant), function(type) {
		  deadlines[type] = defaultDeadline;
      chatFlags[type] = defaultChatFlags;
		});
		var game = {
			private: false,
		  variant: defaultVariant,
			deadlines: deadlines,
			chat_flags: chatFlags,
		};
		this.gameMember = new GameMember({
		  owner: true,
		  game: game
		});
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({}));
		new GameMemberView({ 
		  el: that.$('.create-game'),
			model: that.gameMember,
			button_text: '{{.I "Create" }}',
			button_action: function() {
				that.collection.create(that.gameMember.attributes, {
					success: function() {
						window.session.router.navigate('', { trigger: true });
					},
				});
			},
		}).doRender();
		return that;
	},

});
