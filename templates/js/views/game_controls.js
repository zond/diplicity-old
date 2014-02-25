window.GameControlsView = BaseView.extend({

  template: _.template($('#game_controls_underscore').html()),

	className: "panel panel-default",

	events: {
    "click .view-chat": "viewChat",
    "click .view-orders": "viewOrders",
    "click .view-results": "viewResults",
		"click .commit-phase": "commitPhase",
		"click .uncommit-phase": "uncommitPhase",
	},

	initialize: function(options) {
		this.parentId = options.parentId;
		this.chatMessages = options.chatMessages;
		_.bindAll(this, 'update');
		this.listenTo(this.model, 'change', this.update);
		this.listenTo(this.model, 'reset', this.update);
	},

	commitPhase: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		var that = this;
		RPC('Commit', {
			PhaseId: that.model.get('Phase').Id,
		}, function(error) {
			if (error != null && error != '') {
				logError('While committing', error);
			}
		});
	},

	uncommitPhase: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		var that = this;
		RPC('Uncommit', {
			PhaseId: that.model.get('Phase').Id,
		}, function(error) {
			if (error != null && error != '') {
				logError('While uncommitting', error);
			}
		});
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

	update: function() {
	  var that = this;
		if (that.model.get('Phase') != null) {
			var me = that.model.me();
			if (me != null) {
				if (that.model.get('Phase').Committed[me.Nation]) {
					that.$('a.commit-button').removeClass('commit-phase').addClass('uncommit-phase').attr('title', '{{.I "Uncommit" }}');
					that.$('span.commit-button').removeClass('glyphicon-ok').addClass('glyphicon-remove');
				} else {
					that.$('a.commit-button').removeClass('uncommit-phase').addClass('commit-phase').attr('title', '{{.I "Commit" }}');
					that.$('span.commit-button').removeClass('glyphicon-remove').addClass('glyphicon-ok');
				}
			}
		}
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
