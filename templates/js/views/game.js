window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'provinceClicked');
		this.listenTo(this.model, 'change', this.doRender);
		this.fetch(this.model);
		this.cleaners = [];
	},

	provinceClicked: function(prov) {
	  var that = this;
	  window.wsRPC('ValidOrders', {
		  GameId: this.model.get('Id'),
			Province: prov,
		}, function(result) {
		  that.decide(result);
		});
	},

	decide: function(raw) {
	  var that = this;
		var opts = [];
		var typeMap = {};
		var types = [];
		for (var value in raw) {
			opts.push({
				name: value,
				value: value,
			});
			if (!typeMap[raw[value].Type]) {
				types.push(raw[value].Type);
				typeMap[raw[value].Type] = true;
			}
		}
		if (types.length > 1) {
			logError("Don't know how to decide when having options", raw, "of different types", types);
		} else if (types.length > 0) {
			if (types[0] == "OrderType") {
				new OptionsDialogView({ 
					options: opts,
					selected: function(alternative) {
					  console.log('selected', alternative);
					  that.decide(raw[alternative].Next);
					},
				}).display();
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		}
	},

	renderMap: function(handler) {
	  var that = this;
		var phase = that.model.get('Phase');
		var variant = that.model.get('Variant');
		_.each(that.cleaners, function(cleaner) {
		  cleaner();
		});
		that.cleaners = [];
		var map = dippyMap(that.$('.map'));
	  if (phase != null) {
			map.copySVG(variant + 'Map');
			_.each(phase.Units, function(val, key) {
			  map.addUnit(variant + 'Unit' + val.Type, key, variantColor(variant, val.Nation));
			});
			_.each(variantColorizableProvincesMap[variant], function(key) {
				if (phase.SupplyCenters[key] == null) {
					map.hideProvince(key);
				} else {
					map.colorProvince(key, variantColor(variant, phase.SupplyCenters[key]));
				}
			});
			_.each(variantClickableProvincesMap[variant], function(key) {
				that.cleaners.push(map.addClickListener(key, function(ev) {
				  handler(key);
				}));
			});
			map.showProvinces();
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
