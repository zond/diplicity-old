window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

  events: {
		"click .chat-channel-name": "showMessages",
	},

	initialize: function(options) {
	  var that = this;
	  that.members = options.members;
		that.nMembers = 0;
		that.name = _.map(that.members, function(x, id) {
		  that.nMembers++;
			return id;
		}).sort().join(".");
		that.nameId = that.name.replace(/\./g, "_");
		that.title = _.map(that.members, function(x, id) {
		  var memb = that.model.member(id);
			if (memb == null) {
			  return '{{.I "Anonymous" }}';
			}
			if (that.model.get('State') == {{.GameState "Created"}}) {
				return memb.describe();
			} else {
				return memb.Nation;
			}
		}).sort().join(", ");
		that.listenTo(that.model, 'change', that.updateUnseen);
		that.listenTo(that.model, 'reset', that.updateUnseen);
	},

	showMessages: function(ev) {
		var that = this;
	  new ChatMessagesView({
			el: $('.game-control-container'),
		  collection: that.collection,
			model: that.model,
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
			name: that.name,
			nameId: that.nameId,
			title: that.title,
		}));
		that.updateUnseen();
		return that;
	},

});
