window.GameControlsView = BaseView.extend({

  template: _.template($('#game_controls_underscore').html()),

	className: "panel panel-default",

	initialize: function(options) {
		this.parentId = options.parentId;
		this.listenTo(this.model, 'change', this.doRender);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  parentId: that.parentId,
			model: that.model,
		}));
    return that;
	},
});
