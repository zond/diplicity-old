window.GameResultsView = BaseView.extend({

  template: _.template($('#game_results_underscore').html()),

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		return that;
	},

});
