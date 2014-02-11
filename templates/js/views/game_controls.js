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
		this.listenTo(this.model, 'change', this.doRender);
	},

  viewChat: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.$('.game-controls .panel-body').html(new GameChatView().render().el);
		this.$('.game-controls').collapse('show')
	},

  viewResults: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.$('.game-controls .panel-body').html(new GameResultsView().render().el);
		this.$('.game-controls').collapse('show')
	},

  viewOrders: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.$('.game-controls .panel-body').html(new GameOrdersView().render().el);
		this.$('.game-controls').collapse('show')
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  parentId: that.parentId,
			model: that.model,
		}));
		that.$('.game-controls .panel-body').html(new GameChatView().render().el);
    return that;
	},
});
