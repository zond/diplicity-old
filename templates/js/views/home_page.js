window.HomePageView = Backbone.View.extend({

  template: _.template($('#home_page_underscore').html()),


	initialize: function(options) {
	},

  render: function() {
		var that = this;
		this.$el.html(this.template({}));
		this.$el.trigger('create');
		this.delegateEvents();
		return this;
	},

});
