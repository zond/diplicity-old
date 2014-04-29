window.MapView = BaseView.extend({

  template: _.template($('#map_underscore').html()),

	initialize: function(options) {
	  var that = this;
		that.variant = options.variant;
		that.map = null;
		that.decisionCleaners = [];
		that.data = that.parseData();
	},

  unitReg: /^u(.{3,3})$/,

  parseData: function() {
		var that = this;
		var data = {
		  Units: {},
			Dislodgeds: {},
			Orders: {},
			SupplyCenters: {},
		};
		var params = {};
		parts = window.location.href.split('?');
		if (parts.length > 1) {
      _.each(parts[1].split('&'), function(param) {
				var parts = param.split('=');
				params[parts[0]] = parts[1];
			});
		}
		for (var key in params) {
      var val = params[key];
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
				case 's':
					data.SupplyCenters[key.substring(1,4)] = that.expand('nation',val)
					break;
			}
		}
		for (var key in params) {
			var val = params[key];
			switch (key.substring(0,1)) {
				case 'o':
					var unit = data.Units[key.substring(1,4)];
					if (unit != null) {
						var nat = unit.Nation;
						if (data.Orders[nat] == null) {
							data.Orders[nat] = {};
						}
						data.Orders[nat][key.substring(1,4)] = _.map(val.split(','), function(part) {
							return that.expand('orderType', part);
						});
					}
					break;
			}
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
					case 'V':
						return 'MoveViaConvoy';
					case 'S':
						return 'Support';
					case 'C':
						return 'Convoy';
					case 'H':
						return 'Hold';
					case 'D':
						return 'Disband';
					case 'B':
						return 'Build';
				}
				break;
		}
		return s;
	},

	cleanDecision: function() {
	  var that = this;
		_.each(that.decisionCleaners, function(cleaner) {
			cleaner();
		});
		that.decisionCleaners = [];
	},

	selectProvince: function(callback) {
		var that = this;
		_.each(variantSelectableProvincesMap[that.variant], function(prov) {
			that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
				that.cleanDecision();
				callback(prov);
			}, {
				nohighlight: true,
			}));
		});
	},

	selectUnitType: function(selected, cancelled) {
		var that = this;
		new OptionsDialogView({
      options: [
			  {
					name: '{{.I "Army" }}',
				  value: 'Army',
				},
			  {
					name: '{{.I "Fleet" }}',
				  value: 'Fleet',
				},
			],
		  selected: selected,
			cancelled: cancelled,
		}).display();
	},

	selectNation: function(selected, cancelled) {
		var that = this;
		new OptionsDialogView({
			options: [
			  {
					name: '{{.I "Austria" }}',
				  value: 'Austria',
				},
			  {
					name: '{{.I "England" }}',
				  value: 'England',
				},
			  {
					name: '{{.I "France" }}',
				  value: 'France',
				},
			  {
					name: '{{.I "Germany" }}',
				  value: 'Germany',
				},
			  {
					name: '{{.I "Italy" }}',
				  value: 'Italy',
				},
			  {
					name: '{{.I "Russia" }}',
				  value: 'Russia',
				},
			  {
					name: '{{.I "Turkey" }}',
				  value: 'Turkey',
				},
			],
			selected: selected,
			cancelled: cancelled,
		}).display();
	},

	decide: function() {
	  var that = this;
		that.selectProvince(function(prov) {
			var options = [
    	    {
    				name: '{{.I "Unit" }}',
    	      value: 'unit',
    			},
    	    {
    				name: '{{.I "Dislodged" }}',
    	      value: 'dislodged',
    			},
    	  ];
		  if (that.data.Units[prov] != null) {
				options.push({
    				name: '{{.I "Order" }}',
    	      value: 'order',
    		});
			}
			if (variantSupplyCenterMap[that.variant][prov]) {
				options.push({
					name: '{{.I "Supply center" }}',
			    value: 'sc',
				});
			}
    	new OptionsDialogView({ 
    		options: options,
    	  selected: function(alternative) {
    			that.cleanDecision();
    			switch (alternative) {
						case 'unit':
							that.selectNation(function(nat) {
								that.selectUnitType(function(typ) {
									that.data.Units[prov] = {
										Type: typ,
									  Nation: nat,
									};
									that.render();
								}, function() {
									that.decide();
								});
							}, function() {
								that.decide();
							});
							break;
    				case 'order':
      				new OptionsDialogView({
                options: [
      					  {
      						  name: '{{.I "Move" }}',
                    value: 'Move',
      						},
      					  {
      						  name: '{{.I "Move via convoy" }}',
                    value: 'MoveViaConvoy',
      						},
      					  {
      						  name: '{{.I "Support" }}',
                    value: 'Support',
      						},
      					  {
      						  name: '{{.I "Convoy" }}',
                    value: 'Convoy',
      						},
      					  {
      						  name: '{{.I "Hold" }}',
                    value: 'Hold',
      						},
      					  {
      						  name: '{{.I "Disband" }}',
                    value: 'Disband',
      						},
      					  {
      						  name: '{{.I "Build" }}',
                    value: 'Build',
      						},
      					  {
      						  name: '{{.I "Cancel" }}',
                    value: 'cancel',
      						},
      					],
      					selected: function(alternative) {
      						switch (alternative) {
      							case 'Move':
											that.selectProvince(function(to) {
												var orders = that.data.Orders[that.data.Units[prov].Nation];
												if (orders == null) {
													orders = {};
													that.data.Orders[that.data.Units[prov].Nation] = orders;
												}
												orders[prov] = ['Move', to];
                        that.render();
											});
											break;
      							case 'MoveViaConvoy':
											break;
      							case 'Support':
											break;
      							case 'Convoy':
											break;
      							case 'Hold':
											break;
      							case 'Disband':
											break;
      							case 'Build':
											break;
										case 'cancel':
											break;
      						}
      					},
      					cancelled: function() {
      						that.cleanDecision();
      						that.decide();
      					},
      				}).display();
    					break;
    				case 'dislodged':
    					break;
    				case 'sc':
    					break;
    			}
    		},
    		cancelled: function() {
				  that.decide();
    		},
    	}).display();
		});
	},

  render: function() {
		var that = this;
		that.$el.html(that.template({}));
		that.map = dippyMap(that.$('.map'));
		that.map.copySVG(that.variant + 'Map');
		panZoom('.map');
		for (var prov in that.data.Units) {
			var unit = that.data.Units[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantColor(that.variant, unit.Nation));
		}
		for (var prov in that.data.Dislodgeds) {
			var unit = that.data.Dislodgeds[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantColor(that.variant, unit.Nation), true);
		}
		for (var nation in that.data.Orders) {
			for (var source in that.data.Orders[nation]) {
				that.map.addOrder([source].concat(that.data.Orders[nation][source]), that.variant, nation);
			}
		}
		_.each(variantColorizableProvincesMap[that.variant], function(prov) {
			if (that.data.SupplyCenters[prov] == null) {
				that.map.hideProvince(prov);
			} else {
				that.map.colorProvince(prov, variantColor(that.variant, that.data.SupplyCenters[prov]));
			}
		});
		that.map.showProvinces();
		that.decide();
		return that;
	},

});
