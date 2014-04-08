window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

  events: {
		"click .chat-channel-name": "showMessages",
	},

	initialize: function(options) {
	  var that = this;
	  that.members = options.members;
		that.nMembers = _.size(that.members);
		that.name = ChatMessage.channelIdFor(that.members);
		that.title = ChatMessage.channelTitleFor(that.model, that.members);
		that.listenTo(that.model, 'change', that.updateUnseen);
		that.listenTo(that.model, 'reset', that.updateUnseen);
	},

	showMessages: function(ev) {
		ev.preventDefault();
		var that = this;
	  new ChatMessagesView({
			el: $('.game-control-container'),
		  collection: that.collection,
			model: that.model,
			members: that.members,
		}).doRender();
	},

	updateUnseen: function(ev) {
	  var that = this;
		if (that.model.get('UnseenMessages') != null) {
			var unseen = that.model.get('UnseenMessages')[that.name];
			if (unseen > 0) {
				that.$('.unseen-messages').text(unseen);
				that.$('.unseen-messages').show();
			} else {
				that.$('.unseen-messages').hide();
			}
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
			model: that.model,
			title: that.title,
		}));
		that.updateUnseen();
		return that;
	},

});
