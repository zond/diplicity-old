window.GameMemberView = BaseView.extend({

  template: _.template($('#game_member_underscore').html()),

  events: {
		"change .game-private": "changePrivate",
    "click .game-member-button": "buttonAction",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'save', 'updatePrivate');
		this.model.bind('saveme', this.save);
		this.model.bind('change', this.updatePrivate);
		this.button_text = options.button_text;
		this.button_action = options.button_action;
	},

  buttonAction: function(ev) {
	  ev.preventDefault();
		this.button_action();
	},

	onClose: function() {
	  this.model.unbind('saveme', this.save);
		this.model.unbind('change', this.updatePrivate);
	},

  changePrivate: function(ev) {
	  this.model.get('Game').private = $(ev.target).val() == 'true';
		this.model.trigger('change');
		this.model.trigger('saveme');
	},

  save: function() {
		if (!this.model.isNew()) {
			this.model.save();
		}
	},

	updatePrivate: function() {
		this.$('select.game-private').val(this.model.get('Game').private ? 'true' : 'false');
		this.$('select.game-private').slider().slider('refresh');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  model: that.model,
			owner: that.model.get('Owner'),
			button_text: that.button_text,
		}));
		_.each(variants(), function(variant) {
		  if (variant.id == that.model.get('Game').Variant) {
				that.$('select.create-game-variant').append('<option value="{0}" selected="selected">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			} else {
				that.$('select.create-game-variant').append('<option value="{0}">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			}
		});
		_.each(phaseTypes(that.model.get('Game').Variant), function(type) {
			that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				owner: that.model.get('Owner'),
				game: that.model.get('Game'),
				gameMember: that.model,
			}).doRender().el);
		});
		that.updatePrivate();
		that.$el.trigger('create');
		return that;
	},

});
