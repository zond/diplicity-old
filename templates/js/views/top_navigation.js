window.TopNavigationView = BaseView.extend({

  template: _.template($('#top_navigation_underscore').html()),

	initialize: function(options) {
		this.listenTo(window.session.user, 'change', this.doRender);
		this.online = false;
	},

	online: function(online) {
	  this.online = online;
		this.updateOfflineTag();
	},

  updateOfflineTag: function() {
		if (this.online) {
			this.$('.offline-tag').hide();
		} else {
			this.$('.offline-tag').show();
		}
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		that.updateOfflineTag();
		return that;
	},

});
