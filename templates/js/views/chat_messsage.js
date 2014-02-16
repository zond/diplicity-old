window.ChatMessageView = BaseView.extend({

  template: _.template($('#chat_message_underscore').html()),

  tagName: 'tr',

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  model: that.model,
		}));
		return that;
	},

});
