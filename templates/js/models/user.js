window.User = Backbone.Model.extend({

	url: '/user',

	loggedIn: function() {
		return this.get('email') != null && this.get('email') != '';
	},
});

