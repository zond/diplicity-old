window.BottomNavigationView = BaseView.extend({

  template: _.template($('#bottom_navigation_underscore').html()),

	initialize: function(options) {
		this.buttons = options.buttons;
		this.percent = 100;
	},

	navLinks: function(b) {
	  this.buttons = b;
		this.doRender();
	},

	showPercent: function(num) {
	  console.log('appcache percent', num);
	  this.percent = num;
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
		if (that.percent < 100) {
      that.$('.appcache-progress').show();
		} else {
      that.$('.appcache-progress').hide();
		}
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  percent: that.percent,
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
		$('body').css('margin-bottom', that.$el.height());
		return that;
	},

});
