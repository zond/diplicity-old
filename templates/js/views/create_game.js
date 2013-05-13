window.CreateGameView = Backbone.View.extend({

  template: _.template($('#create_game_underscore').html()),

  events: {
	  "click .create-game-button": "createGame",
		"change .create-game-variant": "changeVariant",
	},

	changeVariant: function(ev) {
	  this.gameMember.get('game').variant = $(ev.target).val();
		this.gameMember.trigger('change');
	},

	createGame: function(ev) {
	  this.collection.create(this.gameMember.attributes);
		$.mobile.changePage('#home');
	},

	initialize: function(options) {
	  _.bindAll(this, 'render');
		var deadlines = {};
		_.each(phaseTypes(defaultVariant), function(type) {
		  deadlines[type] = defaultDeadline;
		});
		var game = {
		  variant: defaultVariant,
			deadlines: deadlines,
		};
		this.gameMember = new GameMember({
		  game: game
		});
	},

  render: function() {
		var that = this;
		this.$el.html(this.template({}));
		_.each(variants(), function(variant) {
		  if (variant.id == that.gameMember.get('game').variant) {
				that.$('select.create-game-variant').append('<option value="{0}" selected="selected">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			} else {
				that.$('select.create-game-variant').append('<option value="{0}">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			}
		});
		this.deadlines = {};
		_.each(phaseTypes(this.$('.create-game-variant').val()), function(type) {
		  that.$('.deadlines').append(new DeadlineSelectView({
				phaseType: type,
				game: that.gameMember.get('game'),
			}).render().el);
		});
		this.$el.trigger('create');
		this.delegateEvents();
		return this;
	},

});
