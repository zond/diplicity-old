window.PreferencesAllocationDialogView = BaseView.extend({

  template: _.template($('#preferences_allocation_dialog_underscore').html()),

	id: 'preferences_allocation_dialog',

  events: {
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.gameState = options.gameState;
	},

	display: function() {
		$('body').append(this.doRender().el);
		$.mobile.changePage("#preferences_allocation_dialog", { role: "dialog" });
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		_.each(variantNations(that.gameState.get('Variant')), function(nation) {
      that.$('.preferences_list').append('<li>' + nation + '</li>');
		});
		return that;
	},

});
