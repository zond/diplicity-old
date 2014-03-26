
function resizeMap() {
	$('.map-container').height($(window).height() - 35);
}

$(window).on('orientationchange', resizeMap);
$(window).on('resize', resizeMap);

window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  var that = this;
		that.originalModel = that.model;
		that.chatParticipants = options.chatParticipants;
		that.sub();
		that.listenTo(window.session.user, 'change', that.doRender);
		that.chatMessages = new ChatMessages([], { url: '/games/' + that.model.get('Id') + '/messages' });
		that.fetch(that.chatMessages);
		that.lastPhaseOrdinal = 0;
		if (that.model.get('Phase') != null) {
		  that.lastPhaseOrdinal = that.model.get('Phase').Ordinal;
		}
		that.fetch(that.model);
		that.decision = null;
		that.decisionFor = null;
		that.decisionCleaners = null;
		that.map = null;
		that.possibleSources = null;
		that.renderedChildren = false;
	},

	phaseForward: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		if (this.lastPhaseOrdinal < this.model.get('Phases')) {
			this.unsub();
			this.model = new GameState({
				Id: this.originalModel.get('Id'),
			}, {
				url: '/games/' + this.originalModel.get('Id') + '/' + (this.lastPhaseOrdinal + 1),
			});
			this.sub();
		}
	},

	lastPhase: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		if (this.lastPhaseOrdinal != this.model.get('Phases')) {
			this.unsub();
			this.model = this.originalModel;
			this.sub();
			this.model.trigger('reset');
		}
	},

	phaseBack: function(ev) {
	  ev.preventDefault();
		ev.stopPropagation();
		if (this.lastPhaseOrdinal > 0) {
			this.unsub();
			this.model = new GameState({
				Id: this.originalModel.get('Id'),
			}, {
				url: '/games/' + this.originalModel.get('Id') + '/' + (this.lastPhaseOrdinal - 1),
			});
			this.sub();
		}
	},

	unsub: function() {
    if (this.model != null) {
		  if (this.model != this.originalModel) {
				this.model.close();
			}
			this.stopListening(this.model);
		}
	},

	sub: function() {
		this.listenTo(this.model, 'change', this.update);
		this.listenTo(this.model, 'reset', this.update);
	  if (this.model != this.originalModel) {
		  this.model.fetch();
		}
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
				dialogOptions.push({
					name: '{{.I "Cancel" }}',
					value: '{{.I "Cancel" }}',
				});
				new OptionsDialogView({ 
					options: dialogOptions,
					selected: function(alternative) {
					  if (alternative == '{{.I "Cancel" }}') {
						  var toCancel = [that.decisionFor];
              var split = that.decisionFor.split("/");
							if (split.length == 2) {
							  toCancel.push(split[0])
							}
							_.each(toCancel, function(provToCancel) {
								RPC('SetOrder', {
									GameId: that.model.get('Id'),
									Order: [provToCancel],
								}, function(error) {
									if (error != null && error != '') {
										logError('While setting order', [provToCancel], error);
									}
									toCancel.pop();
									if (toCancel.length == 0) {
										that.decision = null;
										that.addClickableProvinces();
									}
								});
							});
						} else {
							that.decision.push(alternative);
							that.decide(raw[alternative].Next);
						}
					},
					cancelled: function() {
					  that.addClickableProvinces();
					},
				}).display();
			} else if (types[0] == "UnitType") {
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
			} else if (types[0] == "SrcProvince") {
			  that.decision.unshift(opts[0]);
				that.decide(raw[opts[0]].Next);
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		} else {
		  var decision = that.decision;
			RPC('SetOrder', {
				GameId: that.model.get('Id'),
				Order: decision,
			}, function(error) {
			  if (error != null && error != '') {
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
				RPC('GetValidOrders', {
					GameId: that.model.get('Id'),
					Province: prov,
				}, function(result) {
					that.decision = [];
					that.decisionFor = prov;
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
		panZoom('.map');
		for (var prov in phase.Units) {
			var unit = phase.Units[prov];
			that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantColor(variant, unit.Nation));
		}
		for (var prov in phase.Dislodgeds) {
			var unit = phase.Dislodgeds[prov];
			that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantColor(variant, unit.Nation), true);
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
		if (that.map != null) {
			if (that.model.get('Members') != null) {
			  if (!that.renderedChildren) {
					that.stateView = new GameStateView({ 
						parentId: 'current-game',
						play_state: true,
						editable: false,
						model: that.model,
						el: that.$('.game-state-container'),
					}).doRender();
					that.controlsView = new GameControlsView({
						parentId: 'current-game',
						model: that.model,
						chatMessages: that.chatMessages,
						chatParticipants: that.chatParticipants,
						el: that.$('.game-controls-container'),
						gameView: that,
					}).doRender();
					that.renderedChildren = true;
				}
				if (that.model.get('Phase') != null) {
					if (that.model.get('Phase').Ordinal != that.lastPhaseOrdinal) {
						that.lastPhaseOrdinal = that.model.get('Phase').Ordinal;
						that.possibleSources = null;
					}
					if (that.model.get('Phase').Ordinal < that.model.get('Phases')) {
					  that.controlsView.$('.later-phase').removeAttr('disabled');
					} else {
					  that.controlsView.$('.later-phase').attr('disabled', 'disabled');
					}
					if (that.model.get('Phase').Ordinal > 0) {
					  that.controlsView.$('.previous-phase').removeAttr('disabled');
					} else {
					  that.controlsView.$('.previous-phase').attr('disabled', 'disabled');
					}
				  if (that.model.get('Phase').Ordinal == that.model.get('Phases')) {
						var me = that.model.me();
						if (me != null && that.possibleSources == null) {
							RPC('GetPossibleSources', {
								GameId: that.model.get('Id'),
							}, function(data) {
								that.possibleSources = data;
								that.addClickableProvinces();
							});
						}
					}
					if (that.$('.map').length > 0) {
						that.renderMap();
					}
				} else {
					var variant = that.model.get('Variant');
					that.map.copySVG(variant + 'Map');
					panZoom('.map');
				}
			}
			resizeMap();
		}
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
