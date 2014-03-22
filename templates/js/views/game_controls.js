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
	  this.chatParticipants = options.chatParticipants;
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

	handleClick: function(ev, view) {
		if (ev != null) {
		  ev.preventDefault();
			if (this.currentView != view) {
				ev.stopPropagation();
			}
		}
		this.$('.game-controls-container').hide();
    this.$('.game-' + view + '-container').show();
		this.$('.game-controls').collapse('show')
		this.currentView = view;
	},

  viewChat: function(ev) {
	  var that = this;
		that.gameChatView.doRender();
		that.handleClick(ev, 'chat');
	},

  viewResults: function(ev) {
	  var that = this;
		that.gameResultsView.doRender();
		that.handleClick(ev, 'results');
	},

  viewOrders: function(ev) {
	  var that = this;
		that.gameOrdersView.doRender();
		that.handleClick(ev, 'orders');
	},

	update: function() {
	  var that = this;
		if (that.model.get('Phase') != null) {
		  if (that.chatParticipants != null) {
				that.viewChat();
				that.gameChatView.ensureChannel(_.inject(that.chatParticipants.split("-"), function(sum, nat) {
				  sum[nat] = true;
					return sum
				}, {}));
				that.gameChatView.$('.channel-' + that.chatParticipants).collapse('show');
				that.gameChatView.$('.chevron-' + that.chatParticipants).removeClass('glyphicon-chevron-right').addClass('glyphicon-chevron-down');
				that.chatParticipants = null;
			}
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
		that.gameChatView = new GameChatView({
		  chatParticipants: that.chatParticipants,
			model: that.model,
			collection: that.chatMessages,
			el: that.$('.game-chat-container'),
		});
		that.gameResultsView = new GameResultsView({
			el: that.$('.game-results-container'),
		});
		that.gameOrdersView = new GameOrdersView({
			el: that.$('.game-orders-container'),
		  model: that.model,
		});
    return that;
	},
});
