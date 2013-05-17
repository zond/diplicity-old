window.MenuView = BaseView.extend({

  template: _.template($('#menu_underscore').html()),

  render: function() {
		this.$el.html(this.template({ }));
		return this;
	},

});
