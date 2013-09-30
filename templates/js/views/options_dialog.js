window.OptionsDialogView = BaseView.extend({

  template: _.template($('#options_dialog_underscore').html()),

	className: 'modal fade',

  events: {
		"hidden.bs.modal": "hide",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.options = options.options;
		this.title = options.title;
	},

	hide: function() {
		this.clean(true);
	},

	display: function() {
		$('body').append(this.doRender().el);
		this.$el.modal('show');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  title: that.title,
		}));
		_.each(that.options, function(opt) {
		  that.$('.options-list').append(new OptionView({
			  option: opt,
			}).doRender().el);
		});
		return that;
	},

});
