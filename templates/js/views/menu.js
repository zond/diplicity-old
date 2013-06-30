window.MenuView = BaseView.extend({

  template: _.template($('#menu_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
	  window.session.user.bind('change', this.doRender);
	},

	onClose: function() {
		window.session.user.unbind('change', this.doRender);
	},

  render: function() {
		this.$el.html(this.template({ 
		  user: window.session.user,
		}));
		return this;
	},

});
