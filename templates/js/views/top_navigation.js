window.TopNavigationView = BaseView.extend({

  template: _.template($('#top_navigation_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		return that;
	},

});
