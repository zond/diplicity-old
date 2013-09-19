window.User = Backbone.Model.extend({

  idAttribute: "Id",

  localStorage: true,

	url: '/user',

	loggedIn: function() {
		return this.get('Email') != null && this.get('Email') != '';
	},
});

