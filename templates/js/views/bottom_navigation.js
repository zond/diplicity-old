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
	  var that = this;
	  _.each(that.buttons, function(row) {
		  _.each(row, function(b) {
			  if (b.el != null) {
					if (b.activate == null) {
						if (b.url == window.session.active_url || b.url == '/' + window.session.active_url) {
							b.el.addClass('btn-primary');
						} else {
							b.el.removeClass('btn-primary');
						}
					} else {
						if (b.activate()) {
							b.el.addClass('btn-primary');
						} else {
							b.el.removeClass('btn-primary');
						}
					}
				}
			});
		});
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		_.each(that.buttons, function(row) {
		  var rowEl = $('<div></div>');
			that.$('.buttons').append(rowEl);
		  _.each(row, function(b) {
			  var but;
			  if (b.click == null) {
					but = $('<a href="' + b.url + '" class="btn navigate">' + b.label + '</a>');
				} else {
					but = $('<a href="' + b.url + '" class="btn">' + b.label + '</a>');
					but.on('click', b.click);
				}
				b.el = but;
				rowEl.append(but);
			});
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
