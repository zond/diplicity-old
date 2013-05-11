window.GameMemberView = Backbone.View.extend({

  tagName: 'select',

  template: _.template($('#game_member_underscore').html()),

  render: function() {
    this.$el.html(this.template({
		  model: this.model,
		}));
		return this;
	},

});
