window.ChatChannelView = BaseView.extend({

  className: 'btn btn-default btn-block',

	tagName: 'button',

  template: _.template($('#chat_channel_underscore').html()),

	events: {
	  "click": 'showMessages',
	},

	initialize: function(options) {
	  this.channel = options.channel;
		this.title = options.title;
		this.name = options.name;
		this.game = options.game;
	},

  showMessages: function(ev) {
	  ev.preventDefault();
		var that = this;
		new ChatMessagesView({
		  el: $('#chats-slider'),
			channel: that.channel,
			title: that.title,
			name: that.name,
			game: that.game,
			collection: that.collection,
		}).doRender();
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
