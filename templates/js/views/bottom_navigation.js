window.BottomNavigationView = BaseView.extend({

  template: _.template($('#bottom_navigation_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.buttons = options.buttons;
	},

	navLinks: function(b) {
	  this.buttons = b;
		this.doRender();
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		_.each(that.buttons, function(b) {
		  that.$('.buttons').append('<a href="' + b.url + '" class="btn navigate">' + b.label + '</a>');
		});
		return that;
	},

});
