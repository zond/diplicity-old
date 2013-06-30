window.User = Backbone.Model.extend({

	url: '/user',

	loggedIn: function() {
		return this.get('Email') != null && this.get('Email') != '';
	},
});

