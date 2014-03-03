window.BottomNavigationView = BaseView.extend({

  template: _.template($('#bottom_navigation_underscore').html()),

	initialize: function(options) {
		this.buttons = options.buttons;
	},

	navLinks: function(b) {
	  this.buttons = b;
		this.doRender();
	},
	
	update: function() {
		this.$('a').removeClass('btn-primary');
		this.$('a[href="' + window.session.active_url + '"]').addClass('btn-primary');
		this.$('a[href="/' + window.session.active_url + '"]').addClass('btn-primary');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		_.each(that.buttons, function(b) {
		  that.$('.buttons').append('<a href="' + b.url + '" class="btn navigate">' + b.label + '</a>');
		});
		that.update();
		if (that.buttons.length == 0) {
		  $('body').removeClass('bottom-lift');
		} else {
		  $('body').addClass('bottom-lift');
		}
		return that;
	},

});
