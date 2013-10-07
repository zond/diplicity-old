window.OptionsDialogView = BaseView.extend({

  template: _.template($('#options_dialog_underscore').html()),

	className: 'modal fade',

  events: {
		"hidden.bs.modal": "hide",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'hideAndSelect');
		this.options = options.options;
		this.title = options.title;
		this.selected = options.selected;
		this.cancelled = options.cancelled;
		this.selection = null;
	},

	hide: function() {
		this.clean(true);
	  if (this.selection != null) {
		  this.selected(this.selection);
		} else {
		  if (this.cancelled != null) {
				this.cancelled();
			}
		}
	},

	display: function() {
		$('body').append(this.doRender().el);
		this.$el.modal('show');
	},

  hideAndSelect: function(alternative) {
	  this.selection = alternative;
		this.$el.modal('hide');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  title: that.title,
		}));
		_.each(that.options, function(opt) {
		  that.$('.options-list').append(new OptionView({
			  option: opt,
				selected: that.hideAndSelect,
			}).doRender().el);
		});
		return that;
	},

});
