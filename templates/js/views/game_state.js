window.GameStateView = BaseView.extend({

  template: _.template($('#game_state_underscore').html()),

  events: {
		"change .game-private": "changePrivate",
    "click .game-member-button": "buttonAction",
		"change select.create-game-allocation-method": "changeAllocationMethod",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'save', 'updatePrivate');
		this.model.bind('saveme', this.save);
		this.model.bind('change', this.updatePrivate);
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
		this.model.trigger('change');
		this.model.trigger('saveme');
	},

	onClose: function() {
	  this.model.unbind('saveme', this.save);
		this.model.unbind('change', this.updatePrivate);
	},

  changePrivate: function(ev) {
	  this.model.set('Private', $(ev.target).val() == 'true');
		this.model.trigger('change');
		this.model.trigger('saveme');
	},

  save: function() {
		if (!this.model.isNew()) {
			this.model.save();
		}
	},

	updatePrivate: function() {
		this.$('select.game-private').val(this.model.get('Private') ? 'true' : 'false');
		this.$('select.game-private').slider().slider('refresh');
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
		that.updatePrivate();
		that.$el.trigger('create');
		return that;
	},

});
