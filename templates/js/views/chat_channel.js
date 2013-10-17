window.ChatChannelView = BaseView.extend({

  className: 'panel panel-default',

  template: _.template($('#chat_channel_underscore').html()),

	initialize: function(options) {
	  this.channel = options.channel;
		this.title = options.title;
		this.name = options.name;
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  title: that.title,
			name: that.name,
		}));
		return that;
	},

});
