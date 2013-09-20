window.CreateGameView = BaseView.extend({

  template: _.template($('#create_game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		var deadlines = {};
		var chatFlags = {};
		_.each(phaseTypes(defaultVariant), function(type) {
		  deadlines[type] = defaultDeadline;
      chatFlags[type] = defaultChatFlags;
		});
		var member = {
		  User: btoa(window.session.user.get('Email')),
		};
		this.gameState = new GameState({
		  Member: member,
			Private: false,
		  Variant: defaultVariant,
			Deadlines: deadlines,
			ChatFlags: chatFlags,
			AllocationMethod: defaultAllocationMethod,
		});
		this.gameState.url = '/games';
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		if (window.session.user.loggedIn()) {
			new GameStateView({ 
				el: that.$('.create-game'),
				editable: true,
				model: that.gameState,
				button_text: '{{.I "Create" }}',
				button_action: function() {
				  if (that.gameState.get('AllocationMethod') == 'preferences') {
            new PreferencesAllocationDialogView({ gameState: that.gameState }).display();
					} else {
						that.gameState.save(null, {
							success: function() {
								window.session.router.navigate('', { trigger: true });
							},
						});
					}
				},
			}).doRender();
		}
		return that;
	},

});
