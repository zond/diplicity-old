window.ChatFlagView = BaseView.extend({

  template: _.template($('#chat_flag_underscore').html()),

	className: "form-group",

	events: {
	  "click .chat-flag": "changeChatFlag",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.opt = options.opt;
		this.gameState = options.gameState;
		this.phaseType = options.phaseType;
	},

	changeChatFlag: function(ev) {
	  ev.preventDefault();
		this.gameState.get('ChatFlags')[this.phaseType] ^= this.opt.id;
		this.gameState.trigger('change');
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
	    opt: that.opt,
			gameState: that.gameState,
			phaseType: that.phaseType,
		}));
		return that;
	},

});
