window.SettingsView = BaseView.extend({

  template: _.template($('#settings_underscore').html()),

	events: {
	  "change .user-nickname": "changeNickname",
	  "click .save-button": "saveSettings",
	},

	initialize: function(options) {
		this.listenTo(window.session.user, 'change', this.doRender);
	},

	changeNickname: function(ev) {
	  ev.preventDefault();
		window.session.user.set('Nickname', $(ev.target).val());
	},

	saveSettings: function(ev) {
	  ev.preventDefault();
	  window.session.user.save();
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({
		  model: window.session.user,
		}));
		navLinks(mainButtons);
		return that;
	},

});
