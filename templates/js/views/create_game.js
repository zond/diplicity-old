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
	  this.collection.create(this.gameMember.attributes);
		window.session.router.navigate('', { trigger: true });
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
		that.clean();
		that.$el.html(that.template({}));
		_.each(variants(), function(variant) {
		  if (variant.id == that.gameMember.get('game').variant) {
				that.$('select.create-game-variant').append('<option value="{0}" selected="selected">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			} else {
				that.$('select.create-game-variant').append('<option value="{0}">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			}
		});
		that.deadlines = {};
		_.each(phaseTypes(that.$('.create-game-variant').val()), function(type) {
		  that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				game: that.gameMember.get('game'),
				owner: that.gameMember.get('owner'),
				gameMember: that.gameMember,
			}).doRender().el);
		});
		return that;
	},

});
