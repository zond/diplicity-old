window.MapView = BaseView.extend({

  template: _.template($('#map_underscore').html()),

  events: {
		"click .shortener": "shorten",
	},

	initialize: function(options) {
	  var that = this;
		that.variant = options.variant;
		that.map = null;
		that.decisionCleaners = [];
		that.data = options.data || that.parseData();
	},

  unitReg: /^u(.{3,3})$/,

  shorten: function() {
		var that = this;
		$.ajax('https://www.googleapis.com/urlshortener/v1/url', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify({
				"longUrl": window.location.href,
			}),
			contentType : 'application/json',
			success: function(data) {
				that.$('.shortener').html('<a href="' + data.id + '">' + data.id + '</a>');
				that.$('.shortener').removeClass('glyphicon').removeClass('glyphicon-compressed').removeClass('shortener');	
			},
		});
	},

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
		return data;
	},

  expand: function(cont, s) {
		var that = this;
		switch (cont) {
			case 'nation':
				var found = variantMap[that.variant].NationAbbrevs[s];
				if (found != null) {
					return found;
				}
				break;
			case 'unitType':
				var found = variantMap[that.variant].UnitTypeAbbrevs[s];
				if (found != null) {
					return found;
				}
				break;
			case 'orderType':
				var found = variantMap[that.variant].OrderTypeAbbrevs[s];
				if (found != null) {
					return found;
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
		_.each(variantMap[that.variant].SelectableProvinces, function(prov) {
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
		var options = [
		  {
				name: '{{.I "None" }}',
				value: 'none',
			},
		];
		_.each(variantMap[that.variant].UnitTypes, function(typ) {
			options.push({
				name: {{.I "unit_types" }}[typ],
				value: typ,
			});
		});
		new OptionsDialogView({
      options: options,
		  selected: selected,
			cancelled: cancelled,
		}).display();
	},

	superProv: function(prov) {
		var match = /(.*)\/(.*)/.exec(prov);
		if (match != null) {
			return match[1];
		}
		return prov;
	},

	selectNation: function(selected, cancelled) {
		var that = this;
		var options = [
			{
				name: '{{.I "None" }}',
				value: 'none',
			},
		];
		_.each(variantMap[that.variant].Nations, function(nat) {
			options.push({
				name: {{.I "nations" }}[nat],
				value: nat,
			});
		});
		new OptionsDialogView({
			options: options,
			selected: selected,
			cancelled: cancelled,
		}).display();
	},
	
	selectOrderType: function(selected, cancelled) {
		var that = this;
		var options = [{
			name: '{{.I "None" }}',
			value: 'none',
		}];
		_.each(variantMap[that.variant].OrderTypes, function(typ) {
			options.push({
				name: {{.I "order_types" }}[typ],
				value: typ,
			});
		});
		new OptionsDialogView({
			options: options,
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
		  if (that.data.Units[prov] != null || that.data.Dislodgeds[prov] != null || that.data.SupplyCenters[prov] != null) {
				options.push({
    				name: '{{.I "Order" }}',
    	      value: 'order',
    		});
			}
			if (variantMap[that.variant].SupplyCenters[prov]) {
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
								if (nat == 'none') {
									delete(that.data.Units[prov]);
									that.update();
								} else {
									that.selectUnitType(function(typ) {
										if (typ == 'none') {
											delete(that.data.Units[prov]);
											that.update();
										} else {
											that.data.Units[prov] = {
												Type: typ,
										    Nation: nat,
											};
											that.update();
										}
									}, function() {
										that.decide();
									});
								}
							}, function() {
								that.decide();
							});
							break;
						case 'sc':
							that.selectNation(function(nat) {
								if (nat == 'none') {
									delete(that.data.SupplyCenters[that.superProv(prov)]);
									that.update();
								} else {
									that.data.SupplyCenters[that.superProv(prov)] = nat;
									that.update();
								}
							}, function() {
								that.decide();
							});
							break;
    				case 'order':
							that.selectOrderType(function(alternative) {
      					switch (alternative) {
      						case 'Move':
										that.selectProvince(function(to) {
											var nation = that.data.Units[prov].Nation;
											if (that.data.Dislodgeds[prov] != null) {
												nation = that.data.Dislodgeds[prov].Nation;
											}
											var orders = that.data.Orders[nation];
											if (orders == null) {
												orders = {};
												that.data.Orders[nation] = orders;
											}
											orders[prov] = ['Move', to];
                      that.update();
										});
										break;
      						case 'MoveViaConvoy':
										that.selectProvince(function(to) {
											var orders = that.data.Orders[that.data.Units[prov].Nation];
											if (orders == null) {
												orders = {};
												that.data.Orders[that.data.Units[prov].Nation] = orders;
											}
											orders[prov] = ['Move', to];
                      that.update();
										});
										break;
      						case 'Support':
										that.selectProvince(function(from) {
											that.selectProvince(function(to) {
												var orders = that.data.Orders[that.data.Units[prov].Nation];
												if (orders == null) {
													orders = {};
													that.data.Orders[that.data.Units[prov].Nation] = orders;
												}
												orders[prov] = ['Support', from, to];
												that.update();
											});
										});
										break;
      						case 'Convoy':
										that.selectProvince(function(from) {
											that.selectProvince(function(to) {
												var orders = that.data.Orders[that.data.Units[prov].Nation];
												if (orders == null) {
													orders = {};
													that.data.Orders[that.data.Units[prov].Nation] = orders;
												}
												orders[prov] = ['Convoy', from, to];
												that.update();
											});
										});
										break;
      						case 'Hold':
										var orders = that.data.Orders[that.data.Units[prov].Nation];
										if (orders == null) {
											orders = {};
											that.data.Orders[that.data.Units[prov].Nation] = orders;
										}
										orders[prov] = ['Hold'];
										that.update();
										break;
      						case 'Disband':
										var nation = that.data.Units[prov].Nation;
										if (that.data.Dislodgeds[prov] != null) {
											nation = that.data.Dislodgeds[prov].Nation;
										}
										var orders = that.data.Orders[nation];
										if (orders == null) {
											orders = {};
											that.data.Orders[nation] = orders;
										}
										orders[prov] = ['Disband'];
										that.update();
										break;
      						case 'Build':
										that.selectUnitType(function(typ) {
											if (typ == 'none') {
												delete(that.data.Orders[that.data.SupplyCenters[prov]][prov]);
											} else {
												var orders = that.data.Orders[that.data.SupplyCenters[prov]];
												if (orders == null) {
													orders = {};
													that.data.Orders[that.data.SupplyCenters[prov]] = orders;
												}
												orders[prov] = ['Build', typ];
												that.update();
											}
										}, function() {
											that.decide();
										});
										break;
									case 'none':
										var orders = that.data.Orders[that.data.Units[prov].Nation];
										if (orders != null) {
											delete(orders[prov]);
										}
										orders = that.data.Orders[that.data.Dislodgeds[prov].Nation];
										if (orders != null) {
											delete(orders[prov]);
										}
										orders = that.data.SupplyCenters[that.data.Dislodgeds[prov].Nation];
										if (orders != null) {
											delete(orders[prov]);
										}
										that.update();
										break;
      					}
							},
							function() {
								that.decide();
							});
    					break;
    				case 'dislodged':
							that.selectNation(function(nat) {
								if (nat == 'none') {
									delete(that.data.Dislodgeds[prov]);
									that.update();
								} else {
									that.selectUnitType(function(typ) {
										if (typ == 'none') {
											delete(that.data.Dislodgeds[prov]);
											that.update();
										} else {
											that.data.Dislodgeds[prov] = {
												Type: typ,
										    Nation: nat,
											};
											that.update();
										}
									}, function() {
										that.decide();
									});
								}
							}, function() {
								that.decide();
							});
							break;
					}
    		},
    		cancelled: function() {
				  that.decide();
    		},
    	}).display();
		});
	},

	encodeData: function() {
		return queryEncodePhaseState(this.variant, this.data);
	},

	update: function() {
		var that = this;
		that.map.copySVG(that.variant + 'Map');
		panZoom('.map');
		for (var prov in that.data.Units) {
			var unit = that.data.Units[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantMap[that.variant].Colors[unit.Nation]);
		}
		for (var prov in that.data.Dislodgeds) {
			var unit = that.data.Dislodgeds[prov];
			that.map.addUnit(that.variant + 'Unit' + unit.Type, prov, variantMap[that.variant].Colors[unit.Nation], true);
		}
		for (var nation in that.data.Orders) {
			for (var source in that.data.Orders[nation]) {
				that.map.addOrder([source].concat(that.data.Orders[nation][source]), that.variant, nation);
			}
		}
		_.each(variantMap[that.variant].ColorizableProvinces, function(prov) {
			if (that.data.SupplyCenters[prov] == null) {
				that.map.hideProvince(prov);
			} else {
				that.map.colorProvince(prov, variantMap[that.variant].Colors[that.data.SupplyCenters[prov]]);
			}
		});
		that.map.showProvinces();
		that.decide();
		window.session.router.navigate('/map/' + that.variant + '?' + that.encodeData(), { trigger: false, replace: true });
	},

  render: function() {
		var that = this;
		navLinks([]);
		that.$el.html(that.template({}));
		that.map = dippyMap(that.$('.map'));
		that.update();
		return that;
	},

});
