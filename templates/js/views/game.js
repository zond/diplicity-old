window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'provinceClicked');
		this.listenTo(this.model, 'change', this.doRender);
		this.fetch(this.model);
		this.decision = null;
		this.decisionCleaners = null;
		this.map = null;
	},

	provinceClicked: function(prov) {
		var that = this;
		window.wsRPC('GetValidOrders', {
			GameId: that.model.get('Id'),
			Province: prov,
		}, function(result) {
			that.decide(result);
		});
	},

	shadeProvinces: function(provs) {
		var that = this;
		that.decisionCleaners = [];
	  _.each(provs, function(prov) {
		  that.decisionCleaners.push(that.map.blinkProvince(prov));
		});
	},

	decide: function(raw) {
	  var that = this;
		that.cleanDecision();
		var opts = [];
		var typeMap = {};
		var types = [];
		for (var value in raw) {
			opts.push(value);
			if (!typeMap[raw[value].Type]) {
				types.push(raw[value].Type);
				typeMap[raw[value].Type] = true;
			}
		}
		if (types.length > 1) {
			logError("Don't know how to decide when having options", raw, "of different types", types);
		} else if (types.length > 0) {
			if (types[0] == "OrderType") {
			  var dialogOptions = [];
				_.each(opts, function(opt) {
				  dialogOptions.push({
					  name: opt,
						value: opt,
					});
				});
				new OptionsDialogView({ 
					options: dialogOptions,
					selected: function(alternative) {
					  var provs = [];
						for (var p in raw[alternative].Next) {
						  provs.push(p);
						}
						that.decision = [provs[0], alternative];
					  that.decide(raw[alternative].Next[provs[0]].Next);
					},
				}).display();
			} else if (types[0] == "Province") {
			  _.each(opts, function(prov) {
					that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
					  that.decision.push(prov);
						that.decide(raw[prov].Next);
					}));
				});
			  that.shadeProvinces(opts)
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		} else {
			window.wsRPC('SetOrder', {
				GameId: that.model.get('Id'),
				Order: that.decision,
			});
			that.decision = null;
			that.addClickableProvinces();
		}
	},

	cleanDecision: function() {
	  var that = this;
		_.each(that.decisionCleaners, function(cleaner) {
			cleaner();
		});
		that.decisionCleaners = [];
	},
	
	addClickableProvinces: function() {
	  var that = this;
		var variant = that.model.get('Variant');
		that.cleanDecision();
		_.each(variantMainProvincesMap[variant], function(prov) {
			that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
				that.provinceClicked(prov);
			}));
		});
	},

	renderMap: function(handler) {
	  var that = this;
		var phase = that.model.get('Phase');
		var variant = that.model.get('Variant');
 
		if (that.map != null) {
		  that.$('.map').empty();
		}

		that.map = dippyMap(that.$('.map'));

	  if (phase != null) {
			that.map.copySVG(variant + 'Map');
			for (var prov in phase.Units) {
			  var unit = phase.Units[prov];
			  that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantColor(variant, unit.Nation));
			}
			_.each(variantColorizableProvincesMap[variant], function(prov) {
				if (phase.SupplyCenters[prov] == null) {
					that.map.hideProvince(prov);
				} else {
					that.map.colorProvince(prov, variantColor(variant, phase.SupplyCenters[prov]));
				}
			});
			that.addClickableProvinces();
			that.map.showProvinces();
		}
	},

  render: function() {
		var that = this;
		navLinks([]);
		that.$el.html(that.template({ 
		}));
		if (this.model.get('Members') != null) {
			var state_view = new GameStateView({ 
				parentId: 'current_game',
				play_state: true,
				editable: false,
				model: that.model,
			}).doRender();
			that.$('#current_game').append(state_view.el);
		}
		that.renderMap(this.provinceClicked);
		panZoom('.map');
		return that;
	},

});
