
function resizeMap() {
	$('.map-container').height($(window).height() - 51);
}

$(window).on('orientationchange', resizeMap);
$(window).on('resize', resizeMap);

window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
		this.listenTo(this.model, 'change', this.update);
		this.listenTo(window.session.user, 'change', this.doRender);
		this.stateView = new GameStateView({ 
			parentId: 'current-game',
			play_state: true,
			editable: false,
			model: this.model,
		});
		if (!this.model.isNew()) {
			this.chatMessages = new ChatMessages([], { url: '/games/' + this.model.get('Id') + '/messages' });
			this.fetch(this.chatMessages);
		}
		this.controlsView = new GameControlsView({
		  parentId: 'current-game',
			model: this.model,
			chatMessages: this.chatMessages,
		});
		this.fetch(this.model);
		this.decision = null;
		this.decisionCleaners = null;
		this.map = null;
		this.possibleSources = null;
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
					cancelled: function() {
					  that.addClickableProvinces();
					},
				}).display();
			} else if (types[0] == "Province") {
			  _.each(opts, function(prov) {
					that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
					  that.decision.push(prov);
						that.decide(raw[prov].Next);
					}));
				});
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		} else {
		  var decision = that.decision;
			window.wsRPC('SetOrder', {
				GameId: that.model.get('Id'),
				Order: decision,
			}, function(error) {
			  if (error != '') {
					logError('While setting order', decision, error);
				}
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
		_.each(that.possibleSources, function(prov) {
			that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
				wsRPC('GetValidOrders', {
					GameId: that.model.get('Id'),
					Province: prov,
				}, function(result) {
					that.decide(result);
				});
			}));
		});
	},

	renderMap: function() {
		var that = this;
		var phase = that.model.get('Phase');
		var variant = that.model.get('Variant');

		that.map.copySVG(variant + 'Map');
		for (var prov in phase.Units) {
			var unit = phase.Units[prov];
			that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantColor(variant, unit.Nation));
		}
		for (var nation in phase.Orders) {
			for (var source in phase.Orders[nation]) {
				that.map.addOrder([source].concat(phase.Orders[nation][source]), variant, nation);
			}
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
	},

	update: function() {
	  var that = this;
		if (that.model.get('Members') != null) {
		  if (that.$('#current-game').children().length == 0) {
				that.$('#current-game').append(that.stateView.el);
				that.$('#current-game').append(that.controlsView.el);
			}
			if (that.model.get('Phase') != null) {
				if (that.possibleSources == null) {
					wsRPC('GetPossibleSources', {
						GameId: that.model.get('Id'),
					}, function(data) {
						that.possibleSources = data;
						that.addClickableProvinces();
					});
				}
				if (that.$('.map').length > 0) {
					var hadMap = true;
					if (that.map == null) {
						hadMap = false;
						that.map = dippyMap(that.$('.map'));
					}
					that.renderMap();
					if (!hadMap) {
						panZoom('.map');
					}
				}
			}
		}
		resizeMap();
	},

  render: function() {
		var that = this;
		navLinks([]);
		that.$el.html(that.template({}));
		that.update();
		return that;
	},

});
