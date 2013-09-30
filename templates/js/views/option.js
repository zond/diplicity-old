window.OptionView = BaseView.extend({

  template: _.template($('#option_underscore').html()),

	tagName: 'li',

	className: 'list-group-item',

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.option = options.option;
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  option: that.option,
		}));
		return that;
	},

});
