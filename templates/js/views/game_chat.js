window.GameChatView = BaseView.extend({

  template: _.template($('#game_chat_underscore').html()),

	events: {
	  "click .create-channel-button": "createChannel",
	},

	initialize: function(options) {
	  this.channels = {};
	  this.listenTo(this.collection, 'add', this.addMessage);
	  this.listenTo(this.collection, 'reset', this.loadMessages);
	},

	reloadModel: function() {
	},

	loadMessages: function() {
	  var that = this;
		that.$('#chat-channels').empty();
		that.collection.each(function(message) {
		  if (message.get('SenderId') != null) {
				that.addMessage(message);
			}
		});
	},

	ensureChannel: function(members) {
	  var that = this;
		var channelId = ChatMessage.channelIdFor(members);
		if (that.channels[channelId] == null) {
			var newChannelView = new ChatChannelView({
				collection: that.collection,
				model: that.model,
				members: members,
			}).doRender();
			that.channels[channelId] = newChannelView;
			that.$('#chat-channels').append(newChannelView.el);
		}
		return channelId;
	},

	addMessage: function(message) {
		var that = this;
		var channelId = that.ensureChannel(message.get('RecipientIds'));
	},

	createChannel: function() {
	  var that = this;
	  var memberIds = that.$('.new-channel-members').val().sort();
		memberIds.push(that.model.me().Id);
		if (that.model.allowChatMembers(memberIds.length)) {
		  members = _.inject(memberIds, function(sum, id) {
			  sum[id] = true;
				return sum;
			}, {});
			that.ensureChannel(members);
		} else {
			that.$('.create-channel-container').append('<div class="alert alert-warning fade in">' + 
				'<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' + 
				'<strong>' +
				'{{.I "The game does not allow that particular number of members in a chat channel right now. The only types of chat allowed at the moment are {0}."}}'.format(that.model.describeCurrentChatFlagOptions()) +
				'</strong>' + 
			'</div>');
		}
	},

  render: function() {
	  var that = this;
	  that.channels = {};
    that.$el.html(that.template({
		}));
		var me = that.model.me();
		if (me == null) {
		  that.$('.create-channel-container').hide();
		} else {
			_.each(that.model.members(), function(member) {
			  if (member.Id != me.Id) {
					var opt = $('<option value="' + member.Id + '"></option>');
					if (that.model.get('State') == {{.GameState "Created"}}) {
					  opt.text(member.describe());
					} else {
					  opt.text(member.Nation);
					}
					that.$('.new-channel-members').append(opt);
				}
			});
      var opts = {
				onDropdownHide: function(ev) {
					var el = $(ev.currentTarget);
					el.css('margin-bottom', 0);
				},
				onDropdownShow: function(ev) {
					var el = $(ev.currentTarget);
					el.css('margin-bottom', el.find('.multiselect-container').height());
				},
			};
			that.$('.new-channel-members').multiselect(opts);
		}
		this.loadMessages();
		return that;
	},

});
