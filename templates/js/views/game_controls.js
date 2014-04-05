window.GameControlsView = BaseView.extend({

  template: _.template($('#game_controls_underscore').html()),

	events: {
    "click .view-map": "viewMap",
    "click .view-chat": "viewChat",
    "click .view-orders": "viewOrders",
    "click .view-results": "viewResults",
		"click .commit-phase": "commitPhase",
		"click .uncommit-phase": "uncommitPhase",
		"click .previous-phase": "phaseBack",
		"click .next-phase": "phaseForward",
		"click .last-phase": "lastPhase",
	},

	initialize: function(options) {
	  this.gameView = options.gameView;
	  this.chatParticipants = options.chatParticipants;
		this.parentId = options.parentId;
		this.chatMessages = options.chatMessages;
		_.bindAll(this, 'update');
		this.listenTo(this.chatMessages, 'add', this.update);
		this.listenTo(this.chatMessages, 'reset', this.update);
		this.listenTo(this.model, 'change', this.update);
		this.listenTo(this.model, 'reset', this.update);
		this.timeLeftInterval = null;
		this.lastPhaseOrdinal = 0;
		if (this.model.get('Phase') != null) {
		  this.lastPhaseOrdinal = this.model.get('Phase').Ordinal;
		}
		this.deadline = null;
		this.currentView = null;
	},

	lastPhase: function(ev) {
	  this.gameView.lastPhase(ev);
	},

	phaseForward: function(ev) {
	  this.gameView.phaseForward(ev);
	},

	phaseBack: function(ev) {
	  this.gameView.phaseBack(ev);
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
		var me = that.model.me();
		if (me != null && !me.NoOrders) {
			RPC('Commit', {
				PhaseId: that.model.get('Phase').Id,
			}, function(error) {
				if (error != null && error != '') {
					logError('While committing', error);
				}
			});
		}
	},

	uncommitPhase: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		var that = this;
		var me = that.model.me();
		if (me != null && !me.NoOrders) {
			RPC('Uncommit', {
				PhaseId: that.model.get('Phase').Id,
			}, function(error) {
				if (error != null && error != '') {
					logError('While uncommitting', error);
				}
			});
		}
	},

	viewMap: function(ev) {
		var that = this;
	  if (that.currentView != null) {
			that.currentView.clean();
		}
		that.$('.game-control-container').empty();
	},

  viewChat: function(ev) {
	  var that = this;
		that.currentView = new GameChatView({
		  chatParticipants: that.chatParticipants,
			model: that.gameView.model,
			collection: that.chatMessages,
			el: that.$('.game-control-container'),
		}).doRender();
	},

  viewResults: function(ev) {
	  var that = this;
		that.currentView = new GameResultsView({
			el: that.$('.game-control-container'),
			model: that.gameView.model,
		}).doRender();
	},

  viewOrders: function(ev) {
	  var that = this;
		that.currentView = new GameOrdersView({
			el: that.$('.game-control-container'),
			model: that.gameView.model,
		}).doRender();
	},

	updateTimeLeft: function() {
	  var that = this;
		if (that.deadline == null) {
			that.deadline = new Date(new Date().getTime() + that.model.get('TimeLeft') / 1000000);
		}
		var left = that.deadline.getTime() - new Date().getTime();
		if (left < 0) {
		  that.$('.time-left').hide();
		} else {
		  var secs = left / 1000;
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
		if (that.model.get('State') == {{.GameState "Started"}}) {
			if (that.model.get('Phase').Ordinal != that.lastPhaseOrdinal) {
				that.lastPhaseOrdinal = that.model.get('Phase').Ordinal;
				that.deadline = null;
			}
			that.$('.phase-step').css('visibility', 'visible');
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
				if (me.Committed) {
					that.$('a.commit-button').removeClass('commit-phase').addClass('uncommit-phase').attr('title', '{{.I "Uncommit" }}');
					that.$('span.commit-button').removeClass('glyphicon-ok').addClass('glyphicon-remove');
				} else {
					that.$('a.commit-button').removeClass('uncommit-phase').addClass('commit-phase').attr('title', '{{.I "Commit" }}');
					that.$('span.commit-button').removeClass('glyphicon-remove').addClass('glyphicon-ok');
				}
				if (me.NoOrders) {
				  that.$('a.commit-button').attr('disabled', 'disabled');
				} else {
				  that.$('a.commit-button').removeAttr('disabled');
				}
			} else {
				that.$('.commit-button').css('visibility', 'hidden');
			}
		} else if (that.model.get('State') == {{.GameState "Created"}}) {
			that.$('.commit-button').css('visibility', 'hidden');
			that.$('.phase-step').css('visibility', 'hidden');
			that.$('.view-orders').css('visibility', 'hidden');
			that.$('.view-results').css('visibility', 'hidden');
		} else if (that.model.get('State') == {{.GameState "Ended"}}) {
			that.$('.commit-button').css('display', 'none');
			that.$('.phase-step').css('visibility', 'visible');
			that.$('.view-orders').css('visibility', 'visible');
			that.$('.view-results').css('visibility', 'visible');
			that.$('.end-reason').text('{{.I "Game ended due to {0}"}}'.format(that.model.get('EndReason')));
		}
	},

	reloadModel: function(model) {
    if (this.currentView != null) {
			this.currentView.reloadModel(model);
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  parentId: that.parentId,
			model: that.model,
		}));
		that.update();
		return that;
	},
});
