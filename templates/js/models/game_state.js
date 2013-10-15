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

	currentChatFlags: function() {
	  var that = this;
	  var phase = that.get('phase');
		if (phase == null) {
		  return this.get('ChatFlags')['BeforeGame'] || 0;
		} else { 
			return this.get('ChatFlags')[phase.Type] || 0;
		}
	},

	describePhaseType: function(phaseType) {
	  var that = this;
		var desc = [];
		if (phaseType != 'BeforeGame') {
			desc.push(_.find(deadlineOptions, function(opt) {
				return opt.value == that.get('Deadlines')[phaseType];
			}).name);
		}
		_.each(chatFlagOptions(), function(opt) {
		  if (that.get('ChatFlags')[phaseType] != null && (that.get('ChatFlags')[phaseType] & opt.id) == opt.id) {
			  desc.push(opt.name);
			}
		});
		
	  return desc.join(", ");
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

