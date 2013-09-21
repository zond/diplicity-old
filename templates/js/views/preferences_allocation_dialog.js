window.PreferencesAllocationDialogView = BaseView.extend({

  template: _.template($('#preferences_allocation_dialog_underscore').html()),

	id: 'preferences_allocation_dialog',

  events: {
	  "click .preferences_done": "callDone",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'callDone');
		this.gameState = options.gameState;
		this.done = options.done;
		this.nations = [];
		this.old_page = $.mobile.activePage.attr('id');
		var that = this;
		_.each(variantNations(that.gameState.get('Variant')), function(nation) {
      that.nations.push(nation);
		});
	},

	callDone: function(ev) {
	  ev.preventDefault();
		this.$el.dialog('close');
		this.done(this.nations);
	},

	display: function() {
		$('body').append(this.doRender().el);
		var that = this;
		this.$el.bind('pagehide', function() {
      that.clean();
			that.$el.remove();
		});
		$.mobile.changePage("#preferences_allocation_dialog", { role: 'dialog' });
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		var update_list = null;
		update_list = function() {
			that.cleanChildren();
		  that.$('.preferences_list').empty();
			_.each(that.nations, function(nation) {
				that.$('.preferences_list').append(new PreferredNationView({
					nation: nation,
					action: function() {
						for (var i = 0; i < that.nations.length; i++) {
							var found = that.nations[i];
							if (found == nation) {
								if (i > 0) {
									that.nations[i] = that.nations[i - 1];
									that.nations[i - 1] = found;
								}
								break;
							}
						}
						update_list();
						that.$el.trigger('create');
					},
				}).doRender().el);
			});
		};
		update_list();
		return that;
	},

});
