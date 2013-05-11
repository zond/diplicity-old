window.User = Backbone.Model.extend({
	url: '/user',
	initialize: function(data, options) {

		var that = this;
		this.on('sync', function() {
			if (that.loggedIn()) {
				$('.when-logged-out').css('display', 'none');
				$('.when-logged-in').css('display', 'block');
			} else {
				$('.when-logged-out').css('display', 'block');
				$('.when-logged-in').css('display', 'none');
			}
		});
	},

	loggedIn: function() {
		return this.get('email') != null && this.get('email') != '';
	},
});

