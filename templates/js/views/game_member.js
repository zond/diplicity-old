window.GameMemberView = Backbone.View.extend({

  template: _.template($('#game_member_underscore').html()),

	initialize: function() {
	  _.bindAll(this, 'render');
		this.model.bind('change', this.render);
	},

  render: function() {
	  var that = this;
    this.$el.html(this.template({
		  model: this.model,
		}));
		_.each(phaseTypes(this.model.get('game').variant), function(type) {
		  console.log('rendering', type);
		  that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				game: that.model.get('game'),
			}).render().el);
		});
		return this;
	},

});
