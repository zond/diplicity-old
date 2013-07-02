window.MenuView = BaseView.extend({

  template: _.template($('#menu_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
	  this.listenTo(window.session.user, 'change', this.doRender);
	},

  render: function() {
		this.$el.html(this.template({ 
		  user: window.session.user,
		}));
		return this;
	},

});
