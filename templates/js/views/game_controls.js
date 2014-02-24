window.GameControlsView = BaseView.extend({

  template: _.template($('#game_controls_underscore').html()),

	className: "panel panel-default",

	events: {
    "click .view-chat": "viewChat",
    "click .view-orders": "viewOrders",
    "click .view-results": "viewResults",
	},

	initialize: function(options) {
		this.parentId = options.parentId;
		this.chatMessages = options.chatMessages;
	},

  viewChat: function(ev) {
		this.$('.game-controls .panel-body').html(new GameChatView({
		  model: this.model,
			collection: this.chatMessages,
		}).render().el);
		this.handleClick(ev, 'chat');
	},

	handleClick: function(ev, view) {
		if (ev != null) {
		  ev.preventDefault();
			if (this.currentView != view) {
				ev.stopPropagation();
				this.$('.game-controls').collapse('show')
				this.currentView = view;
			}
		}
	},

  viewResults: function(ev) {
		this.$('.game-controls .panel-body').html(new GameResultsView().render().el);
		this.handleClick(ev, 'results');
	},

  viewOrders: function(ev) {
		this.$('.game-controls .panel-body').html(new GameOrdersView({
		  model: this.model,
		}).render().el);
		this.handleClick(ev, 'orders');
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  parentId: that.parentId,
			model: that.model,
		}));
    return that;
	},
});
