window.GameStateView = BaseView.extend({

  template: _.template($('#game_state_underscore').html()),

	className: "panel panel-default",

  events: {
		"click .game-private": "changePrivate",
		"click .game-ranking": "changeRanking",
    "click .game-state-button": "buttonAction",
		"change .game-allocation-method": "changeAllocationMethod",
		"change .game-variant": "changeVariant",
		"hide.bs.collapse .players": "collapsePlayers",
		"show.bs.collapse .players": "expandPlayers",
		"hide.bs.collapse .game": "collapse",
		"show.bs.collapse .game": "expand",
		"click .game-secret-flag": "changeSecretFlag",
		"click .game-consequence": "changeConsequence",
	},

	initialize: function(options) {
		this.play_state = options.play_state;
		this.editable = options.editable;
		this.expanded = this.editable;
		this.membersExpanded = false;
		this.parentId = options.parentId;
		this.phaseTypeViews = {};
		this.memberViews = {};
		this.reloadModel(this.model);
	},

	reloadModel: function(model) {
	  this.stopListening();
		this.model = model;
		this.listenTo(this.model, 'change', this.doRender);
		this.listenTo(this.model, 'reset', this.doRender);
		this.doRender();
	},

	changeSecretFlag: function(ev) {
	  ev.preventDefault();
		var type = $(ev.target).attr('data-secret-type');
		var flag = parseInt($(ev.target).attr('data-secret-flag'));
		var currently = this.model.get(type);
		this.model.set(type, currently ^ flag);
	},

	changeConsequence: function(ev) {
	  ev.preventDefault();
		var type = $(ev.target).attr('data-consequence-type') + 'Consequences';
		var currently = this.model.get(type);
		var flag = parseInt($(ev.target).attr('data-consequence'));
		this.model.set(type, currently ^ flag);
	},

	collapsePlayers: function(ev) {
	  this.membersExpanded = false;
	},

	expandPlayers: function(ev) {
	  this.membersExpanded = true;
	},

	collapse: function(ev) {
	  this.expanded = false;
	},

	expand: function(ev) {
	  this.expanded = true;
	},

  buttonAction: function(ev) {
		var that = this;
		ev.preventDefault();
		var save_call = function() {
			that.model.save(null, {
				success: function() {
					navigate('/mine/forming');
				},
			});
		};
	  if (that.model.isNew()) {
			if (that.model.get('AllocationMethod') == 'preferences') {
				new PreferencesAllocationDialogView({ 
					gameState: that.model,
					done: function(nations) {
						that.model.get('Members')[0].PreferredNations = nations;
						save_call();
					},
				}).display();
			} else {
				save_call();
			}
		} else {
		  var me = that.model.me();
			if (me == null) {
				that.model.set('Members', [
					{
						UserId: btoa(window.session.user.get('Email')),
						User: {},
					}
				]);
				if (that.model.get('AllocationMethod') == 'preferences') {
					new PreferencesAllocationDialogView({ 
						gameState: that.model,
						done: function(nations) {
							that.model.get('Members')[0].PreferredNations = nations;
							save_call();
						},
					}).display();
				} else {
					save_call();
				}
			} else {
			  that.model.destroy();
			}
		}
	},

	changeVariant: function(ev) {
	  this.model.set('Variant', $(ev.target).val(), { silent: true });
		this.updateDescription();
	},

	changeAllocationMethod: function(ev) {
	  this.model.set('AllocationMethod', $(ev.target).val(), { silent: true });
		this.updateDescription();
	},

  changeRanking: function(ev) {
	  this.model.set('Ranking', $(ev.target).is(':checked'), { silent: true });
		this.updateDescription();
	},

  changePrivate: function(ev) {
	  this.model.set('Private', $(ev.target).is(':checked'), { silent: true });
		this.updateDescription();
	},

	updateDescription: function() {
    this.$('.game-description').text(this.model.describe());
	},

	buttonText: function() {
	  if (this.model.isNew()) {
		  return '{{.I "Create" }}';
		} else {
		  var me = this.model.me();
			if (me == null) {
				return '{{.I "Join" }}';
			} else {
				return '{{.I "Leave" }}';
			}
		}
	},
  render: function() {
	  var that = this;
		var classes = [];
		if (!that.editable) {
		  classes.push('panel-collapse');
			classes.push('collapse');
		}
		if (that.expanded) {
		  classes.push('in');
		}
		var unseen = _.inject(that.model.get('UnseenMessages') || {}, function(sum, num, x) {
			return sum + num;
		}, 0);
    that.$el.html(that.template({
		  classes: classes,
			play_state: that.play_state,
		  parentId: that.parentId,
			membersExpanded: that.membersExpanded,
		  model: that.model,
			editable: that.editable,
			button_text: that.buttonText(),
			unseenMessages: unseen,
		}));
		if (unseen == 0) {
		  that.$('.game-description-container .unseen-messages').hide();
		} else {
		  that.$('.game-description-container .unseen-messages').show();
		}
		_.each(variants, function(variant) {
			that.$('.game-variant').append('<option value="{0}">{1}</option>'.format(variant, variantMap[variant].Name));
		});
		that.$('.game-variant').val(that.model.get('Variant'));
		_.each(allocationMethods, function(meth) {
			that.$('.game-allocation-method').append('<option value="{0}">{1}</option>'.format(meth, allocationMethodMap[meth].Name));
		});
		that.$('.game-allocation-method').val(that.model.get('AllocationMethod'));
		if (that.editable) {
			var newPhaseTypeViews = {};
			_.each(['BeforeGame'].concat(phaseTypes(that.model.get('Variant'))).concat(['AfterGame']), function(phaseType) {
				var phaseTypeView = that.phaseTypeViews[phaseType];
				if (phaseTypeView == null) {
					phaseTypeView = new PhaseTypeView({
						parent: that,
						phaseType: phaseType,
						parentId: 'game_' + that.model.cid + '_phase_types',
						editable: that.editable,
						gameState: that.model,
					}).doRender();
				} else {
					phaseTypeView.doRender();
				}
				that.$('.phase-types').append(phaseTypeView.el);
				newPhaseTypeViews[phaseType] = phaseTypeView;
			});
			for (var phaseType in that.phaseTypes) {
				if (newPhaseTypeViews[phaseType] == null) {
				  that.phaseTypes[phaseType].clean(true);
				}
			}
			that.phaseTypeViews = newPhaseTypeViews;
		} else {
			_.each(['BeforeGame'].concat(phaseTypes(that.model.get('Variant'))).concat(['AfterGame']), function(phaseType) {
				that.$('.phase-types').append('<tr><td>' + {{.I "phase_types" }}[phaseType] + '</td><td>' + that.model.describePhaseType(phaseType) + '</td></tr>');
			});
		}
		var newMemberViews = {};
		_.each(that.model.members(), function(member) {
			var memberView = that.memberViews[member.Id];
			if (memberView == null) {
				memberView = new GameMemberView({
					member: member,
				}).doRender();
			} else {
			  memberView.member = member;
			  memberView.doRender();
			}
			that.$('.game-players').append(memberView.el);
			newMemberViews[member.Id] = memberView;
		});
		for (var memberId in that.memberViews) {
		  if (newMemberViews[memberId] == null) {
			  that.memberViews[memberId].clean(true);
			}
		}
		that.memberViews = newMemberViews;
		if (that.model.get('State') == {{.GameState "Started"}}) {
			var minutesLeft = that.model.get('TimeLeft') / (1000000000 * 60);
			var part = 1 - (minutesLeft / that.model.get('Deadlines')[that.model.get('Phase').Type]);
			that.$('.urgency-bar').css('width', ($(window).width() - 4) * part);
			that.$('.urgency-bar').show();
	    var me = that.model.me();
	    if (me.Committed) {
			  that.$('.urgency-bar').addClass('urgency-bar-green').removeClass('urgency-bar-red');
			} else {
			  that.$('.urgency-bar').addClass('urgency-bar-red').removeClass('urgency-bar-green');
			}	
		} else {
			that.$('.urgency-bar').hide();
		}
		return that;
	},

});
