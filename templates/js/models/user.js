window.User = Backbone.Model.extend({
	url: '/user',
	initialize: function(data, options) {
	  var gameMembers = options.gameMembers;

		var that = this;
		this.bind('change', function() {
			if (that.get('email') != null && that.get('email') != '') {
				gameMembers.fetch();
			}
		});
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

