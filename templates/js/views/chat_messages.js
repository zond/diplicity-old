window.ChatMessagesView = BaseView.extend({

  template: _.template($('#chat_messages_underscore').html()),

	events: {
	  "click .back-button": "showChannels",
		"change .message": "sendMessage",
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

	sendMessage: function(ev) {
	  ev.preventDefault();
		console.log('gonna send');
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
    if ((that.game.currentChatFlags() & {{.ChatFlag "Grey" }}) == {{.ChatFlag "Grey"}}) {
			that.$('.sender-selection').append('<option value="Anonymous">{{.I "Anonymous" }}</option>');
		}
		_.each(variantNations(that.game.get('Variant')), function(nation) {
			if (that.game.me().Nation == nation || (that.game.currentChatFlags() & {{.ChatFlag "Black" }}) == {{.ChatFlag "Black"}}) {
			  that.$('.sender-selection').append('<option ' + (that.game.me().Nation == nation ? 'selected="selected" ' : '') + 'value="' + nation + '">' + {{.I "nations" }}[nation] + '</option>');
			}
		});
		return that;
	},

});
