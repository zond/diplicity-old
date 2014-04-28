window.MapView = BaseView.extend({

  template: _.template($('#map_underscore').html()),

	initialize: function(options) {
	  var that = this;
		that.variant = options.variant;
		that.map = null;
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({}));
		that.map = dippyMap(that.$('.map'));
		that.map.copySVG(that.variant + 'Map');
		panZoom('.map');
		_.each(variantColorizableProvincesMap[that.variant], function(prov) {
			that.map.hideProvince(prov);
		});
		that.map.showProvinces();
		return that;
	},

});
