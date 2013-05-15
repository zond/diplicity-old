window.GameMemberView = Backbone.View.extend({

  template: _.template($('#game_member_underscore').html()),

  events: {
		"change .game-private": "changePrivate",
		"click .leave-game": "leaveGame",
	},

	initialize: function(options) {
	  _.bindAll(this, 'render', 'save', 'onClose', 'updatePrivate');
		this.model.bind('saveme', this.save);
		this.model.bind('change', this.updatePrivate);
		options.parent.children.push(this);
		this.createdAt = new Date().getTime();
		this.children = [];
	},

	leaveGame: function(ev) {
	  this.model.destroy();
	},

  changePrivate: function(ev) {
	  this.model.get('game').private = $(ev.target).val() == 'true';
		this.model.trigger('change');
		this.model.trigger('saveme');
	},

	onClose: function() {
	  this.model.unbind('saveme', this.save);
	},

  save: function() {
	  this.model.save();
	},

	clean: function() {
	  _.each(this.children, function(child) {
		  child.onClose();
		});
		this.children = [];
	},

	updatePrivate: function() {
		this.$('select.game-private').val(this.model.get('game').private ? 'true' : 'false');
		this.$('select.game-private').slider().slider('refresh');
	},

  render: function() {
	  var that = this;
	  that.clean();
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
				parent: that,
			}).render().el);
		});
		that.updatePrivate();
		that.$el.trigger('create');
		return that;
	},

});
