window.User = Backbone.Model.extend({
	url: '/user',
	initialize: function() {
		var that = this;
		this.on('sync', function() {
			if (that.get('email') == null || that.get('email') == '') {
				$('.login-button').css('display', 'block');
				$('.logout-button').css('display', 'none');
			} else {
				$('.login-button').css('display', 'none');
				$('.logout-button').css('display', 'block');
			}
		});
	},
});

