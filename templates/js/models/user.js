window.User = Backbone.Model.extend({
	url: '/user',
	initialize: function() {
		var that = this;
		this.on('sync', function() {
			if (that.get('email') == null || that.get('email') == '') {
				$('.when-logged-out').css('display', 'block');
				$('.when-logged-in').css('display', 'none');
			} else {
				$('.when-logged-out').css('display', 'none');
				$('.when-logged-in').css('display', 'block');
			}
		});
	},
});

