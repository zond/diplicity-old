window.OptionView = BaseView.extend({

  template: _.template($('#option_underscore').html()),

	tagName: 'li',

	className: 'list-group-item',

  events: {
	  "click .btn": "click",
	},

	initialize: function(options) {
		this.option = options.option;
		this.selected = options.selected;
	},

  click: function(ev) {
	  ev.preventDefault();
		this.selected($(ev.target).attr('data-value'));
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  option: that.option,
		}));
		return that;
	},

});
