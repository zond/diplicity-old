window.BottomNavigationView = BaseView.extend({

  template: _.template($('#bottom_navigation_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		return that;
	},

});
