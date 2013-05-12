window.DeadlineSelectView = Backbone.View.extend({

  template: _.template($('#deadline_select_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'render');
		this.phaseType = options.phaseType;
	},

  render: function() {
		this.$el.html(this.template({
		  phaseType: this.phaseType,
		}));
		return this;
	},

});
