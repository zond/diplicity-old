window.CreateGameView = BaseView.extend({

  template: _.template($('#create_game_underscore').html()),

  events: {
	  "click a.create-game-button": "createGame",
		"change select.create-game-variant": "changeVariant",
		"change select.create-game-private": "changePrivate",
	},

  changePrivate: function(ev) {
	  this.gameMember.get('game').private = $(ev.target).val() == 'true';
		this.gameMember.trigger('change');
	},

	changeVariant: function(ev) {
	  this.gameMember.get('game').variant = $(ev.target).val();
		this.gameMember.trigger('change');
	},

	createGame: function(ev) {
	  ev.preventDefault();
	  this.collection.create(this.gameMember.attributes, {
		  success: function() {
				window.session.router.navigate('', { trigger: true });
			},
		});
	},

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
		}).doRender();
		return that;
	},

});
