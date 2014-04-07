window.ChatMessagesView = BaseView.extend({

  template: _.template($('#chat_messages_underscore').html()),

  initialize: function(opts) {
		this.name = opts.name;
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		that.collection.each(function(msg) {
			if (that.name == ChatMessage.channelIdFor(msg.get('RecipientIds'))) {
				that.$('.channel-messages').prepend(new ChatMessageView({
					model: msg,
					game: that.model,
				}).doRender().el);
			}
		});
		var maxHeight = $(window).height() - $('#top-navigation').height() - $('#current-game').height() - 27;
		if (that.$('.channel-messages').height() > maxHeight) {
			that.$('.channel-messages').height(maxHeight);
		}
		if ($('.game-control-container').height() > maxHeight) {
			$('.game-control-container').height(maxHeight);
		}
		return that;
	},
});
