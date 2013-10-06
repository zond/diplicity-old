window.GameStateView = BaseView.extend({

  template: _.template($('#game_state_underscore').html()),

	className: "panel panel-default",

  events: {
		"click .game-private": "changePrivate",
		"click .game-secret-email": "changeSecretEmail",
		"click .game-secret-nickname": "changeSecretNickname",
		"click .game-secret-nation": "changeSecretNation",
    "click .game-state-button": "buttonAction",
		"change .game-allocation-method": "changeAllocationMethod",
		"change .game-variant": "changeVariant",
		"hide.bs.collapse .players": "collapsePlayers",
		"show.bs.collapse .players": "expandPlayers",
		"hide.bs.collapse .game": "collapse",
		"show.bs.collapse .game": "expand",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.play_state = options.play_state;
		this.button_text = options.button_text;
		this.button_action = options.button_action;
		this.editable = options.editable;
		this.expanded = this.editable;
		this.membersExpanded = false;
		this.parentId = options.parentId;
		this.phaseTypeViews = {};
		this.memberViews = {};
		this.listenTo(this.model, 'change', this.doRender);
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
	  ev.preventDefault();
		this.button_action();
	},

	changeVariant: function(ev) {
	  this.model.set('Variant', $(ev.target).val(), { silent: true });
		this.updateDescription();
	},

	changeAllocationMethod: function(ev) {
	  this.model.set('AllocationMethod', $(ev.target).val(), { silent: true });
		this.updateDescription();
	},

  changePrivate: function(ev) {
	  this.model.set('Private', $(ev.target).val() == 'true', { silent: true });
		this.updateDescription();
	},

  changeSecretEmail: function(ev) {
	  this.model.set('SecretEmail', $(ev.target).val() == 'true', { silent: true });
		this.updateDescription();
	},

  changeSecretNickname: function(ev) {
	  this.model.set('SecretNickname', $(ev.target).val() == 'true', { silent: true });
		this.updateDescription();
	},

  changeSecretNation: function(ev) {
	  this.model.set('SecretNation', $(ev.target).val() == 'true', { silent: true });
		this.updateDescription();
	},

	updateDescription: function() {
    this.$('.game-description').text(this.model.describe());
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
    that.$el.html(that.template({
		  classes: classes,
			play_state: that.play_state,
		  parentId: that.parentId,
			membersExpanded: that.membersExpanded,
		  model: that.model,
			editable: that.editable,
			button_text: that.button_text,
		}));
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
			_.each(phaseTypes(that.model.get('Variant')), function(phaseType) {
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
			_.each(phaseTypes(that.model.get('Variant')), function(phaseType) {
				that.$('.phase-types').prepend('<tr><td>' + phaseType + '</td><td>' + that.model.describePhaseType(phaseType) + '</td></tr>');
			});
		}
		var newMemberViews = {};
		_.each(that.model.get('Members'), function(member) {
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
		return that;
	},

});
