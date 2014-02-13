window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

	className: "panel panel-default",

	initialize: function(options) {
	  this.members = options.members;
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  members: that.members,
			model: that.model,
		}));
		return that;
	},

});
