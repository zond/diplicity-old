window.GameStateView = BaseView.extend({

  template: _.template($('#game_state_underscore').html()),

  events: {
		"change .game-private": "changePrivate",
		"change .game-secret-email": "changeSecretEmail",
		"change .game-secret-nickname": "changeSecretNickname",
		"change .game-secret-nation": "changeSecretNation",
    "click .game-member-button": "buttonAction",
		"change select.create-game-allocation-method": "changeAllocationMethod",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.button_text = options.button_text;
		this.button_action = options.button_action;
		this.editable = options.editable;
	},

  buttonAction: function(ev) {
	  ev.preventDefault();
		this.button_action();
	},

	changeAllocationMethod: function(ev) {
	  this.model.set('AllocationMethod', $(ev.target).val());
		this.update();
	},

  changePrivate: function(ev) {
	  this.model.set('Private', $(ev.target).val() == 'true');
		this.update();
	},

  changeSecretEmail: function(ev) {
	  this.model.set('SecretEmail', $(ev.target).val() == 'true');
	},

  changeSecretNickname: function(ev) {
	  this.model.set('SecretNickname', $(ev.target).val() == 'true');
	},

  changeSecretNation: function(ev) {
	  this.model.set('SecretNation', $(ev.target).val() == 'true');
	},

	update: function() {
	  this.$('.description').text(this.model.describe());
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  model: that.model,
			editable: that.editable,
			button_text: that.button_text,
		}));
		_.each(variants(), function(variant) {
		  if (variant.id == that.model.get('Variant')) {
				that.$('select.create-game-variant').append('<option value="{0}" selected="selected">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			} else {
				that.$('select.create-game-variant').append('<option value="{0}">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
			}
		});
		_.each(allocationMethods(), function(meth) {
		  if (meth.id == that.model.get('AllocationMethod')) {
				that.$('select.create-game-allocation-method').append('<option value="{0}" selected="selected">{{.I "Allocation method"}}: {1}</option>'.format(meth.id, meth.name));
			} else {
				that.$('select.create-game-allocation-method').append('<option value="{0}">{{.I "Allocation method"}}: {1}</option>'.format(meth.id, meth.name));
			}
		});
		_.each(phaseTypes(that.model.get('Variant')), function(type) {
			that.$('.phase-types').append(new PhaseTypeView({
				phaseType: type,
				editable: that.editable,
				gameState: that.model,
			}).doRender().el);
		});
		_.each(that.model.get('Members'), function(member) {
		  console.log('rendering', member);
		  that.$('.member-list').append(new GameMemberView({
			  member: member,
			}).doRender().el);
		});
		return that;
	},

});
