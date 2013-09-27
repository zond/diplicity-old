window.GameState = Backbone.Model.extend({

  urlRoot: '/games',

  localStorage: function() {
	  return this.get('Id') != null;
	},
	
	me: function() {
	  if (window.session.user.get('Email') == null) {
		  return null;
		}
	  return _.find(this.get('Members'), function(member) {
		  return member.UserId == btoa(window.session.user.get('Email'));
		});
	},

	describePhaseType: function(phaseType) {
	  var that = this;
		var desc = [];
	  desc.push(_.find(deadlineOptions, function(opt) {
			return opt.value == that.get('Deadlines')[phaseType];
		}).name);
		_.each(chatFlagOptions(), function(opt) {
		  if (that.get('ChatFlags')[phaseType] != null && (that.get('ChatFlags')[phaseType] & opt.id) == opt.id) {
			  desc.push(opt.name);
			}
		});
		
	  return desc.join(", ");
	},

	render: function(destSel) {
	  var that = this;
		var phase = that.get('Phase');
		var variant = that.get('Variant');
	  if (phase != null) {
			$(destSel).copySVG(variant + 'Map');
			_.each(phase.Units, function(val, key) {
			  $(destSel + ' svg').addUnit(variant + 'Unit' + val.Type, key, variantColor(variant, val.Nation));
			});
			_.each(variantProvincesMap[variant], function(key) {
				if (phase.SupplyCenters[key] == null) {
					$(destSel + ' svg').hideProvince(key);
				} else {
					$(destSel + ' svg').colorProvince(key, variantColor(variant, phase.SupplyCenters[key]));
				}
			});
			$(destSel + ' svg').find('#provinces')[0].removeAttribute('style');
		}
	},

	describe: function() {
	  var that = this;
		var nationInfo = allocationMethodName(that.get('AllocationMethod'));
		var member = that.me();
		if (member != null && member.Nation != null && member.Nation != '') {
		  var nationInfo = {{.I "nations" }}[member.Nation];
		}
		var phase = that.get('Phase');
		var phaseInfo = '{{.I "Forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.Season], phase.Year, {{.I "phase_types"}}[phase.Type]);
		}
		var info = [nationInfo, phaseInfo, variantName(that.get('Variant'))];
		var lastDeadline = null;
		_.each(phaseTypes(that.get('Variant')), function(phaseType) {
		  var thisDeadline = that.get('Deadlines')[phaseType];
			if (thisDeadline != lastDeadline) {
				info.push(deadlineName(thisDeadline));
				lastDeadline = thisDeadline;
			}
		});
		return info.join(", ");
	},
});

