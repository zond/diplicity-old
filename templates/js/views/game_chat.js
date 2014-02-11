window.GameChatView = BaseView.extend({

  template: _.template($('#game_chat_underscore').html()),

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		return that;
	},

});
