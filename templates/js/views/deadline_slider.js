window.DeadlineSliderView = Backbone.View.extend({

  template: _.template($('#deadline_slider_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'render');
		this.phaseType = options.phaseType;
	},

  render: function() {
		this.$el.html(this.template({}));
		return this;
	},

});
