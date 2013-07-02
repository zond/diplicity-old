window.User = Backbone.Model.extend({

  localStorage: true,

	url: '/user',

	loggedIn: function() {
		return this.get('Email') != null && this.get('Email') != '';
	},
});

