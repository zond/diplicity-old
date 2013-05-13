window.GameMemberView = Backbone.View.extend({

  template: _.template($('#game_member_underscore').html()),

	initialize: function() {
	  _.bindAll(this, 'render');
	},

  render: function() {
	  var that = this;
    this.$el.html(this.template({
		  model: this.model,
			owner: that.model.get('owner'),
		}));
		_.each(phaseTypes(this.model.get('game').variant), function(type) {
		  that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				owner: that.model.get('owner'),
				game: that.model.get('game'),
				gameMember: that.model,
			}).render().el);
		});
		this.$('.game-private').val(this.model.get('game').private ? 'true' : 'false');
		this.$el.trigger('create');
		return this;
	},

});
