window.GameControlsView = BaseView.extend({

  template: _.template($('#game_controls_underscore').html()),

	events: {
    "click .view-chat": "viewChat",
    "click .view-orders": "viewOrders",
    "click .view-results": "viewResults",
		"click .commit-phase": "commitPhase",
		"click .uncommit-phase": "uncommitPhase",
		"hide.bs.collapse .game-controls": "hideControls",
		"click .previous-phase": "phaseBack",
		"click .next-phase": "phaseForward",
		"click .last-phase": "lastPhase",
	},

	initialize: function(options) {
		this.originalModel = this.model;
	  this.chatParticipants = options.chatParticipants;
		this.parentId = options.parentId;
		this.chatMessages = options.chatMessages;
		_.bindAll(this, 'update');
		this.listenTo(this.chatMessages, 'add', this.update);
		this.listenTo(this.chatMessages, 'reset', this.update);
		this.sub();
		this.timeLeftInterval = null;
	},

	phaseForward: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.unsub();
		this.model = new GameState({
			Id: this.originalModel.get('Id'),
		});
		this.sub();
	},

	unsub: function() {
    if (this.model != null) {
		  if (this.model != this.originalModel) {
				this.model.close();
			}
			this.stopListening(this.model);
		}
	},

	sub: function() {
	  if (this.model != this.originalModel) {
		  this.model.fetch();
		}
		this.listenTo(this.model, 'change', this.update);
		this.listenTo(this.model, 'reset', this.update);
	},

	lastPhase: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.unsub();
		this.model = this.originalModel;
		this.sub();
	},

	phaseBack: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		this.unsub();
		this.model = new GameState({
			Id: this.originalModel.get('Id'),
		});
		this.sub();
	},

	hideControls: function(ev) {
		window.session.router.navigate('/games/' + this.model.get('Id'), { trigger: false });
	  if ($(ev.target).hasClass('game-controls')) {
			this.$('.channel').each(function(x, el) {
				$(el).collapse('hide');
			});
		}
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

	updateTimeLeft: function() {
	  var that = this;
		var left = that.model.get('TimeLeft');
		if (left < 0) {
		  that.$('.time-left').hide();
		} else {
		  var secs = left / 1000000000;
			if (secs > 3600 * 24) {
			  that.$('.time-left').text('{{.I "{0}d" }}'.format(parseInt(secs / (3600 * 24))));
			} else if (secs > 3600) {
			  that.$('.time-left').text('{{.I "{0}h" }}'.format(parseInt(secs / 3600)));
			} else if (secs > 60) {
			  that.$('.time-left').text('{{.I "{0}m" }}'.format(parseInt(secs / 60)));
			} else {
			  that.$('.time-left').text('{{.I "{0}s" }}'.format(parseInt(secs)));
			}
			that.$('.time-left').show();
		}
	},

	update: function() {
	  var that = this;
		if (that.model.get('Members') != null) {
			var unseen = _.inject(that.model.get('UnseenMessages') || {}, function(sum, num, x) {
				return sum + num;
			}, 0);
			if (unseen > 0) {
			  that.$('.game-controls-left .unseen-messages').text(unseen);
			  that.$('.game-controls-left .unseen-messages').show();
			} else {
			  that.$('.game-controls-left .unseen-messages').hide();
			}
			if (that.chatParticipants != null) {
				that.viewChat();
				that.gameChatView.ensureChannel(_.inject(that.chatParticipants.split("."), function(sum, nat) {
					sum[nat] = true;
					return sum
				}, {}));
				that.gameChatView.$('.channel-' + that.chatParticipants.replace(/\./g, "_")).collapse('show');
				that.gameChatView.$('.chevron-' + that.chatParticipants.replace(/\./g, "_")).removeClass('glyphicon-chevron-right').addClass('glyphicon-chevron-down');
				that.chatParticipants = null;
			}
		}
		if (that.model.get('Phase') != null) {
			that.updateTimeLeft();
		  if (that.timeLeftInterval != null) {
			  window.clearInterval(that.timeLeftInterval);
			}
			that.timeLeftInterval = window.setInterval(function() {
			  that.updateTimeLeft();
			}, 1000);
			that.$('.view-orders').css('visibility', 'visible');
			that.$('.view-results').css('visibility', 'visible');
			var me = that.model.me();
			if (me != null) {
				that.$('.commit-button').css('visibility', 'visible');
				if (that.model.get('Phase').Committed[me.Nation]) {
					that.$('a.commit-button').removeClass('commit-phase').addClass('uncommit-phase').attr('title', '{{.I "Uncommit" }}');
					that.$('span.commit-button').removeClass('glyphicon-ok').addClass('glyphicon-remove');
				} else {
					that.$('a.commit-button').removeClass('uncommit-phase').addClass('commit-phase').attr('title', '{{.I "Commit" }}');
					that.$('span.commit-button').removeClass('glyphicon-remove').addClass('glyphicon-ok');
				}
			} else {
				that.$('.commit-button').css('visibility', 'hidden');
			}
		} else {
			that.$('.commit-button').css('visibility', 'hidden');
			that.$('.view-orders').css('visibility', 'hidden');
			that.$('.view-results').css('visibility', 'hidden');
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
