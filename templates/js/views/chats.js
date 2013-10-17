window.ChatsView = BaseView.extend({

  template: _.template($('#chats_underscore').html()),

	initialize: function(options) {
		this.game = options.game;
		this.listenTo(window.session.user, 'change', this.doRender);
		this.listenTo(this.collection, "sync", this.update);
		this.listenTo(this.collection, "reset", this.update);
		this.listenTo(this.collection, "add", this.update);
		this.listenTo(this.collection, "remove", this.update);
		this.fetch(this.collection);
		this.channelViews = {};
	},

	update: function() {
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		if ((that.game.currentChatFlags() & {{.ChatFlag "Conference" }}) == {{.ChatFlag "Conference" }}) {
		/*
		  that.channelViews['Conference'] = new ChatChannelView({
			  channel: that.game.conferenceChannel(),
			});
		*/
	}
		return that;
	},

});
