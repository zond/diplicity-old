window.CreateGameView = BaseView.extend({

  template: _.template($('#create_game_underscore').html()),

	initialize: function(options) {
		this.listenTo(window.session.user, 'change', this.doRender);
		var deadlines = {};
		var chatFlags = {};
		_.each(phaseTypes(defaultVariant), function(type) {
		  deadlines[type] = defaultDeadline;
      chatFlags[type] = defaultChatFlags;
		});
		chatFlags['BeforeGame'] = defaultChatFlags;
		var member = {
		  UserId: btoa(window.session.user.get('Email')),
			User: {
			},
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
		navLinks(mainButtons);
		that.gameState.get('Members')[0].User = window.session.user.attributes;
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		if (window.session.user.loggedIn()) {
		  var save_call = function() {
				that.gameState.save(null, {
					success: function() {
						navigate('/');
					},
				});
			};
			var state_view = new GameStateView({ 
				parentId: 'create_game',
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
			that.$('#create_game').append(state_view.el);
		}
		return that;
	},

});
