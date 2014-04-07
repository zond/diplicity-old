window.ChatMessagesView = BaseView.extend({

  template: _.template($('#chat_messages_underscore').html()),

  initialize: function(opts) {
		this.name = opts.name;
		this.nMembers = opts.nMembers;
		this.members = opts.members;
		this.listenTo(this.model, 'change', this.doRender);
	},

  events: {
		"click .create-message-button": "createMessage",
    "keyup .new-message-body": "keyup",
	},

  keyup: function(ev) {
	  if (ev.keyCode == 13 && !ev.shiftKey && !ev.altKey) { 
			this.createMessage(ev);
		}
	},

	createMessage: function(ev) {
	  var that = this;
		ev.preventDefault();
		if (that.model.allowChatMembers(that.nMembers)) {
			// i have NO IDEA AT ALL why i have to use this clunky id scheme to find the body, but that.$('.new-message-body').val() never produced anything but ''
			var body = $('.new-message-body').val();
			if (body != '') {
				$('.new-message-body').val('');
				that.collection.create({
					RecipientIds: that.members,
					Body: body,
					GameId: that.model.get('Id'),
				}, { silent: true });
			}
		} else {
			that.$('.channel-messages').prepend('<div class="alert alert-warning fade in">' + 
				'<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' + 
				'<strong>' +
				'{{.I "The game does not allow that particular number of members in a chat channel right now. The only types of chat allowed at the moment are {0}."}}'.format(that.model.describeCurrentChatFlagOptions()) +
				'</strong>' + 
			'</div>');
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		that.collection.each(function(msg) {
			if (msg.get('SenderId') != null && that.name == ChatMessage.channelIdFor(msg.get('RecipientIds'))) {
				that.$('.messages-container').prepend(new ChatMessageView({
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
