window.ChatMessageView = BaseView.extend({

  template: _.template($('#chat_message_underscore').html()),

  tagName: 'tr',

	initialize: function(options) {
	  this.game = options.game;
	},

  render: function() {
	  var that = this;
		console.log('showing', that.model);
		that.$el.html(that.template({
		  model: that.model,
			sender: that.game.member(that.model.get('Sender')),
		}));
		that.$('abbr.timeago').timeago();
		return that;
	},

});
