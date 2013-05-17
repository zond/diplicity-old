window.GameMemberView = BaseView.extend({

  template: _.template($('#game_member_underscore').html()),

  events: {
		"change .game-private": "changePrivate",
		"click .leave-game": "leaveGame",
		"click .join-game": "joinGame",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'save', 'updatePrivate');
		this.model.bind('saveme', this.save);
		this.model.bind('change', this.updatePrivate);
		this.createdAt = new Date().getTime();
		this.onJoin = options.onJoin;
	},

	onClose: function() {
	  this.model.unbind('saveme', this.save);
		this.model.unbind('change', this.updatePrivate);
	},

	leaveGame: function(ev) {
	  ev.preventDefault();
	  this.model.destroy();
	},

	joinGame: function(ev) {
	  ev.preventDefault();
	  var that = this;
	  that.model.save(null, {
    	success: function() {
				if (that.onJoin != null) {
					that.onJoin();
				}
			},
		});
	},

  changePrivate: function(ev) {
	  this.model.get('game').private = $(ev.target).val() == 'true';
		this.model.trigger('change');
		this.model.trigger('saveme');
	},

  save: function() {
	  this.model.save();
	},

	updatePrivate: function() {
		this.$('select.game-private').val(this.model.get('game').private ? 'true' : 'false');
		this.$('select.game-private').slider().slider('refresh');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  model: that.model,
			owner: that.model.get('owner'),
		}));
		_.each(phaseTypes(that.model.get('game').variant), function(type) {
			that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				owner: that.model.get('owner'),
				game: that.model.get('game'),
				gameMember: that.model,
			}).doRender().el);
		});
		that.updatePrivate();
		that.$el.trigger('create');
		return that;
	},

});
