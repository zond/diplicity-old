window.ChatMessageView = BaseView.extend({

  template: _.template($('#chat_message_underscore').html()),

  tagName: 'tr',

	initialize: function(options) {
	  this.game = options.game;
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  model: that.model,
			sender: that.game.member(that.model.get('Sender')),
		}));
		return that;
	},

});
