window.ChatMessagesView = BaseView.extend({

  template: _.template($('#chat_messages_underscore').html()),

	events: {
	  "click .back-button": "showChannels",
	},

	initialize: function(options) {
		this.listenTo(this.collection, "sync", this.doRender);
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
	  this.channel = options.channel;
		this.title = options.title;
		this.name = options.name;
		this.game = options.game;
	},

	showChannels: function(ev) {
	  var that = this;
	  ev.preventDefault();
		new ChatsView({
			el: $('#chats-slider'),
			game: that.game,
			collection: that.collection,
		}).doRender();
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		return that;
	},

});
