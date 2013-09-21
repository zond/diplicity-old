window.PreferredNationView = BaseView.extend({

  template: _.template($('#preferred_nation_underscore').html()),


	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.nation = options.nation;
		this.ordinal = options.ordinal;
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  ordinal: that.ordinal,
		  nation: that.nation,
		}));
		return that;
	},

});
