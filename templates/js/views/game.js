window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'provinceClicked');
		this.listenTo(this.model, 'change', this.doRender);
		this.fetch(this.model);
		this.cleaners = [];
		this.decision = null;
		this.decisionNext = null;
		this.decisionCleaners = null;
		this.map = null;
	},

	provinceClicked: function(prov) {
	  var that = this;
		if (that.decisionNext != null) {
		  if (that.decisionNext[prov] != null) {
				var chosenNext = that.decisionNext[prov].Next;
				that.decision.push(prov);
				that.decisionNext = null;
				that.decide(chosenNext);
			}
		} else {
			window.wsRPC('ValidOrders', {
				GameId: this.model.get('Id'),
				Province: prov,
			}, function(result) {
			  that.decision = [prov];
				that.decide(result);
			});
		}
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
		_.each(that.decisionCleaners, function(cleaner) {
		  cleaner();
		});
		that.decisionCleaners = [];
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
					  that.decision.push(alternative);
					  that.decide(raw[alternative].Next);
					},
				}).display();
			} else if (types[0] == "Province") {
			  that.decisionNext = raw;
			  that.shadeProvinces(opts)
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		} else {
	    console.log('decided', that.decision);
			that.decision = null;
		}
	},

	renderMap: function(handler) {
	  var that = this;
		var phase = that.model.get('Phase');
		var variant = that.model.get('Variant');
 
		// Clean event listeners for old map, if any
		_.each(that.cleaners, function(cleaner) {
		  cleaner();
		});
		// Remove old map, if any
		if (that.map != null) {
		  that.$('.map').empty();
		}

		// Initialize new cleaners and map
		that.cleaners = [];
		that.map = dippyMap(that.$('.map'));

	  if (phase != null) {
			that.map.copySVG(variant + 'Map');
			_.each(phase.Units, function(val, key) {
			  that.map.addUnit(variant + 'Unit' + val.Type, key, variantColor(variant, val.Nation));
			});
			_.each(variantColorizableProvincesMap[variant], function(key) {
				if (phase.SupplyCenters[key] == null) {
					that.map.hideProvince(key);
				} else {
					that.map.colorProvince(key, variantColor(variant, phase.SupplyCenters[key]));
				}
			});
			_.each(variantClickableProvincesMap[variant], function(key) {
				that.cleaners.push(that.map.addClickListener(key, function(ev) {
				  handler(key);
				}));
			});
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
