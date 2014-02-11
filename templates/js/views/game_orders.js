window.GameOrdersView = BaseView.extend({

  template: _.template($('#game_orders_underscore').html()),

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		return that;
	},

});
