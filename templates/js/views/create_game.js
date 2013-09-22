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
		  UserId: btoa(window.session.user.get('Email')),
			User: {},
		};
		this.gameState = new GameState({
		  Members: [member],
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
		  var save_call = function() {
				that.gameState.save(null, {
					success: function() {
						window.session.router.navigate('', { trigger: true });
					},
				});
			};
			new GameStateView({ 
				el: that.$('.create-game'),
				editable: true,
				model: that.gameState,
				button_text: '{{.I "Create" }}',
				button_action: function() {
				  if (that.gameState.get('AllocationMethod') == 'preferences') {
            new PreferencesAllocationDialogView({ 
							gameState: that.gameState,
							done: function(nations) {
								that.gameState.get('Members')[0].PreferredNations = nations;
                save_call();
							},
						}).display();
					} else {
					  save_call();
					}
				},
			}).doRender();
		}
		return that;
	},

});
