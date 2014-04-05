window.ChatMessagesView = BaseView.extend({

  template: _.template($('#chat_messages_underscore').html()),

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		return that;
	},
});
