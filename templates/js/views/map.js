window.MapView = BaseView.extend({

  template: _.template($('#map_underscore').html()),

	initialize: function(options) {
	  var that = this;
		that.variant = options.variant;
		that.map = null;
	},

  unitReg: /^u(.{3,3})$/,

  getData: function() {
		var that = this;
		var data = {
		  Units: {},
			Dislodgeds: {},
			Orders: {},
			SupplyCenters: {},
		};
		parts = window.location.href.split('?');
		if (parts.length > 1) {
      _.each(parts[1].split('&'), function(param) {
				var parts = param.split('=');
				var key = parts[0];
				var val = parts[1];
				switch (key.substring(0,1)) {
					case 'u':
						data.Units[key.substring(1,4)] = {
							Type: that.expand('unitType', val.substring(0,1)),
				      Nation: that.expand('nation', val.substring(1,2)),
						};
						break;
					case 'd':
						data.Dislodgeds[key.substring(1,4)] = {
							Type: that.expand('unitType', val.substring(0,1)),
				      Nation: that.expand('nation', val.substring(1,2)),
						};
						break;
					case 'o':
						var nat = that.expand('nation', key.substring(1,2));
						if (data.Orders[nat] == null) {
							data.Orders[nat] = {};
						}
						data.Orders[nat][key.substring(2,5)] = _.map(val.split(','), function(part) {
							return that.expand('orderType', part);
						});
						break;
					case 's':
						data.SupplyCenters[key.substring(1,4)] = that.expand('nation',val)
						break;
				}
			});
		}
		console.log(data);
		return data;
	},

  expand: function(cont, s) {
		switch (cont) {
			case 'nation':
				switch (s) {
					case 'A':
						return 'Austria';
					case 'E':
						return 'England';
					case 'G':
						return 'Germany';
					case 'I':
						return 'Italy';
					case 'R':
						return 'Russia';
					case 'T':
						return 'Turkey';
					case 'F':
						return 'France';
				}
				break;
			case 'unitType':
				switch (s) {
					case 'A':
						return 'Army';
					case 'F':
						return 'Fleet';
				}
				break;
			case 'orderType':
				switch (s) {
					case 'M':
						return 'Move';
					case 'S':
						return 'Support';
					case 'C':
						return 'Convoy';
					case 'H':
						return 'Hold';
					case 'D':
						return 'Disband';
					case 'R':
						return 'Remove';
					case 'B':
						return 'Build';
				}
				break;
		}
		return s;
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({}));
		that.map = dippyMap(that.$('.map'));
		that.map.copySVG(that.variant + 'Map');
		panZoom('.map');
		var data = that.getData();
		for (var prov in data.Units) {
			var unit = data.Units[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantColor(that.variant, unit.Nation));
		}
		for (var prov in data.Dislodgeds) {
			var unit = data.Dislodgeds[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantColor(that.variant, unit.Nation), true);
		}
		for (var nation in data.Orders) {
			for (var source in data.Orders[nation]) {
				that.map.addOrder([source].concat(data.Orders[nation][source]), that.variant, nation);
			}
		}
		_.each(variantColorizableProvincesMap[that.variant], function(prov) {
			if (data.SupplyCenters[prov] == null) {
				that.map.hideProvince(prov);
			} else {
				that.map.colorProvince(prov, variantColor(that.variant, data.SupplyCenters[prov]));
			}
		});
		that.map.showProvinces();
		return that;
	},

});
