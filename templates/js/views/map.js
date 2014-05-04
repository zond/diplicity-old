window.MapView = BaseView.extend({

  template: _.template($('#map_underscore').html()),

  events: {
		"click .shortener": "shorten",
    "click .season": "selectSeason",
    "click .year": "selectYear",
    "click .phase-type": "selectPhaseType",
	},

	initialize: function(options) {
	  var that = this;
		that.variant = options.variant;
		that.map = null;
		that.decisionCleaners = [];
		that.data = that.parseData(options.href || window.location.href);
	},

  selectYear: function(ev) {
		ev.preventDefault();
		var that = this;
		var options = [
		];
		for (var i = 1901; i < 1936; i++) {
			options.push({
				name: '' + i,
				value: '' + i,
			});
		}
		new OptionsDialogView({
      options: options,
		  selected: function(sel) {
				that.data.Year = sel;
				that.update();
			},
			cancelled: function() {},
		}).display();
	},

	selectPhaseType: function(ev) {
		ev.preventDefault();
		var that = this;
		var options = [
		];
		_.each(variantMap[that.variant].PhaseTypes, function(typ) {
			options.push({
				name: {{.I "phase_types" }}[typ],
				value: typ,
			});
		});
		new OptionsDialogView({
      options: options,
		  selected: function(typ) {
				that.data.Type = typ;
				that.update();
			},
			cancelled: function() {},
		}).display();
	},

	selectSeason: function(ev) {
		ev.preventDefault();
		var that = this;
		var options = [
		];
		_.each(variantMap[that.variant].Seasons, function(season) {
			options.push({
				name: {{.I "seasons" }}[season],
				value: season,
			});
		});
		new OptionsDialogView({
      options: options,
		  selected: function(sel) {
				that.data.Season = sel;
				that.update();
			},
			cancelled: function() {},
		}).display();
	},

  unitReg: /^u(.{3,3})$/,

  shorten: function(ev) {
		ev.preventDefault();
		var that = this;
		$.ajax('https://www.googleapis.com/urlshortener/v1/url', {
			type: 'POST',
			dataType: 'json',
			data: JSON.stringify({
				"longUrl": window.location.href,
			}),
			contentType : 'application/json',
			success: function(data) {
				that.$('.shortener').text(data.id);
				that.$('.shortener').attr('href', data.id);
			},
		});
	},

  parseData: function(href) {
		var that = this;
		var data = {
		  Units: {},
			Dislodgeds: {},
			Orders: {},
			SupplyCenters: {},
      Season: variantMap[that.variant].Seasons[0],
			Year: "1901",
			Type: variantMap[that.variant].PhaseTypes[0],
		};
		var params = {};
		parts = href.split('?');
		if (parts.length > 1) {
      _.each(parts[1].split('&'), function(param) {
				var parts = param.split('=');
				params[parts[0]] = parts[1];
			});
		}
		if (params.year != null) {
			data.Year = parseInt(params.year);
			delete(params.year);
		}
		if (params.season != null) {
			data.Season = params.season;
			delete(params.season);
		}
		if (params.type != null) {
			data.Type = params.type;
			delete(params.type);
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
					var nat = that.expand('nation', key.substring(4,5));
					if (data.Orders[nat] == null) {
						data.Orders[nat] = {};
					}
					data.Orders[nat][key.substring(1,4)] = _.map(val.split(','), function(part) {
						return that.expand('orderType', part);
					});
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
										if (that.data.Units[prov] != null) {
											if (that.data.Orders[that.data.Units[prov].Nation] != null) {
												delete(that.data.Orders[that.data.Units[prov].Nation][prov]);
											}
										}
										if (that.data.Dislodgeds[prov] != null) {
											if (that.data.Orders[that.data.Dislodgeds[prov].Nation] != null) {
												delete(that.data.Orders[that.data.Dislodgeds[prov].Nation][prov]);
											}
										}
										if (that.data.SupplyCenters[prov] != null) {
											delete(that.data.Orders[that.data.SupplyCenters[prov]][prov]);
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
		that.$('.season').text(that.data.Season);
		that.$('.year').text(that.data.Year);
		that.$('.phase-type').text(that.data.Type);
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
