
function resizeMap() {
	$('.map-container').height($(window).height() - 35);
}

$(window).on('orientationchange', resizeMap);
$(window).on('resize', resizeMap);

window.GameView = BaseView.extend({

  template: _.template($('#game_underscore').html()),

	initialize: function(options) {
	  var that = this;
		that.wantedOrdinal = options.ordinal;
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
		that.lastProvinceClicked = null;
		that.decisionCleaners = null;
		that.map = null;
		that.renderedChildren = false;
	},

	phaseForward: function(ev) {
		var that = this;
	  ev.preventDefault();
		ev.stopPropagation();
		if (that.lastPhaseOrdinal < that.model.get('Phases')) {
			that.unsub();
			that.model = new GameState({
				Id: that.originalModel.get('Id'),
			}, {
				url: '/games/' + that.originalModel.get('Id') + '/' + (that.lastPhaseOrdinal + 1),
			});
			if (that.lastPhaseOrdinal < that.model.get('Phases') - 1) {
				window.session.router.navigate("/games/" + that.originalModel.get('Id') + '/' + (that.lastPhaseOrdinal + 1));
			} else {
				window.session.router.navigate("/games/" + that.originalModel.get('Id'));
			}
			that.sub();
			that.controlsView.reloadModel(that.model);
			that.stateView.reloadModel(that.model);
		}
	},

	lastPhase: function(ev) {
		var that = this;
	  ev.preventDefault();
		ev.stopPropagation();
		if (that.lastPhaseOrdinal != that.model.get('Phases')) {
			that.unsub();
			that.model = that.originalModel;
			window.session.router.navigate("/games/" + that.originalModel.get('Id'));
			that.sub();
			that.model.trigger('reset');
			that.controlsView.reloadModel(that.model);
			that.stateView.reloadModel(that.model);
		}
	},

	phaseBack: function(ev) {
		var that = this;
	  ev.preventDefault();
		ev.stopPropagation();
		if (that.lastPhaseOrdinal > 0) {
			that.unsub();
			that.model = new GameState({
				Id: that.originalModel.get('Id'),
			}, {
				url: '/games/' + that.originalModel.get('Id') + '/' + (that.lastPhaseOrdinal - 1),
			});
			window.session.router.navigate("/games/" + that.originalModel.get('Id') + '/' + (that.lastPhaseOrdinal - 1))
			that.sub();
			that.controlsView.reloadModel(that.model);
			that.stateView.reloadModel(that.model);
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
						  var toCancel = [that.lastProvinceClicked];
              var split = that.lastProvinceClicked.split("/");
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
										that.resetDecision();
									}
								});
							});
						} else {
							that.decision.push(alternative);
							that.decide(raw[alternative].Next);
						}
					},
					cancelled: function() {
					  that.resetDecision();
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
					  that.resetDecision();
					},
				}).display();
			} else if (types[0] == "Province") {
			  _.each(opts, function(prov) {
					that.decisionCleaners.push(that.map.addClickListener(prov, function(ev) {
					  if (that.decision.length > 0) {
							that.decision.push(prov);
						}
						that.lastProvinceClicked = prov;
						that.decide(raw[prov].Next);
					}));
				});
			} else if (types[0] == "SrcProvince") {
			  that.decision.unshift(opts[0]);
				that.decide(raw[opts[0]].Next);
			} else {
			  logError("Don't know how to handle options of type", types[0]);
			}
		} else if (that.decision.length > 0) {
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
			that.resetDecision();
		}
	},

	cleanDecision: function() {
	  var that = this;
		_.each(that.decisionCleaners, function(cleaner) {
			cleaner();
		});
		that.decisionCleaners = [];
	},
	
	renderMap: function() {
		var that = this;
		var phase = that.model.get('Phase');
		var variant = that.model.get('Variant');

		that.map.copySVG(variant + 'Map');
		panZoom('.map');
		for (var prov in phase.Units) {
			var unit = phase.Units[prov];
			that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantMap[variant].Colors[unit.Nation]);
		}
		for (var prov in phase.Dislodgeds) {
			var unit = phase.Dislodgeds[prov];
			that.map.addUnit(variant + 'Unit' + unit.Type, prov, variantMap[variant].Colors[unit.Nation], true);
		}
		for (var nation in phase.Orders) {
			for (var source in phase.Orders[nation]) {
				that.map.addOrder([source].concat(phase.Orders[nation][source]), variant, nation);
			}
		}
		_.each(variantMap[variant].ColorizableProvinces, function(prov) {
			if (phase.SupplyCenters[prov] == null) {
				that.map.hideProvince(prov);
			} else {
				that.map.colorProvince(prov, variantMap[variant].Colors[phase.SupplyCenters[prov]]);
			}
		});
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
					that.controlsView.updatePlannerURL(that.model);
					if (that.model.get('Phase').Ordinal != that.lastPhaseOrdinal) {
						that.lastPhaseOrdinal = that.model.get('Phase').Ordinal;
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
					if (that.$('.map').length > 0) {
						that.renderMap();
					}
				  if (that.model.get('Phase').Ordinal == that.model.get('Phases')) {
						that.resetDecision();
					}
					if (that.wantedOrdinal != null) {
						var wanted = that.wantedOrdinal;
						that.wantedOrdinal = null;
						that.unsub();
						that.model = new GameState({
							Id: that.originalModel.get('Id'),
						}, {
							url: '/games/' + that.originalModel.get('Id') + '/' + wanted,
						});
						window.session.router.navigate("/games/" + that.originalModel.get('Id') + '/' + wanted);
						that.sub();
						that.controlsView.reloadModel(that.model);
						that.stateView.reloadModel(that.model);
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

	resetDecision: function() {
		this.decision = [];
		var me = this.model.me();
		if (me != null) {
			this.decide(me.Options);
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
