window.ChatMessageView = BaseView.extend({

  template: _.template($('#chat_message_underscore').html()),

	initialize: function(options) {
	  this.game = options.game;
	},

  render: function() {
	  var that = this;
		var sender = that.game.member(that.model.get('SenderId'));
		var senderName = '{{.I "Anonymous" }}';
		if (sender != null) {
			if (that.game.get('State') == {{.GameState "Created" }}) {
				senderName = sender.shortDescribe();
			} else {
				senderName = sender.Nation;
			}
		}
		that.$el.html(that.template({
		  model: that.model,
			senderName: senderName,
		}));
		return that;
	},

});
