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
		"hide.bs.collapse .game": "collapse",
		"show.bs.collapse .game": "expand",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.button_text = options.button_text;
		this.button_action = options.button_action;
		this.editable = options.editable;
		this.expanded = this.editable;
		this.parentId = options.parentId;
		this.listenTo(this.model, 'change', this.doRender);
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
		  parentId: that.parentId,
		  model: that.model,
			editable: that.editable,
			button_text: that.button_text,
		}));
		_.each(variants(), function(variant) {
			that.$('.game-variant').append('<option value="{0}">{1}</option>'.format(variant.id, variant.name));
		});
		that.$('.game-variant').val(that.model.get('Variant'));
		_.each(allocationMethods(), function(meth) {
			that.$('.game-allocation-method').append('<option value="{0}">{1}</option>'.format(meth.id, meth.name));
		});
		that.$('.game-allocation-method').val(that.model.get('AllocationMethod'));
		_.each(phaseTypes(that.model.get('Variant')), function(phaseType) {
			if (that.editable) {
				that.$('.phase-types').append(new PhaseTypeView({
				  parent: that,
					phaseType: phaseType,
					parentId: (that.model.get('Id') || '') + '_phase_types',
					editable: that.editable,
					gameState: that.model,
				}).doRender().el);
			} else {
				that.$('.phase-types').prepend('<tr><td>' + phaseType + '</td><td>' + that.model.describePhaseType(phaseType) + '</td></tr>');
			}
		});
		_.each(that.model.get('Members'), function(member) {
		  console.log('member', member);
		  that.$('.game-players').append(new GameMemberView({
			  member: member,
			}).doRender().el);
		});
		return that;
	},

});
