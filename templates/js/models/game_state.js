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
		  return member.User.Email == window.session.user.get('Email');
		});
	},

	conferenceChannel: function() {
		var result = {};
		_.each(variantNations(this.get('Variant')), function(nation) {
		  result[nation] = true;
		});
		return result
	},

	appendSecrecy: function(type, phaseType, desc) {
	  var flag = this.get(type);
		if ((flag & secretFlagMap[phaseType]) == secretFlagMap[phaseType]) {
		  desc.push(secrecyTypesMap[type]);
		}
	},

	currentChatFlags: function() {
	  var that = this;
		if (that.get('State') != null) {
			if (that.get('State') == {{.GameState "Created" }}) {
				return that.get('ChatFlags')['BeforeGame'] || 0;
			} else if (that.get('State') == {{.GameState "Ended" }}) {
				return that.get('ChatFlags')['AfterGame'] || 0;
			}
			var phase = that.get('Phase');
			return that.get('ChatFlags')[phase.Type] || 0;
		}
		return 0;
	},

	describePhaseType: function(phaseType) {
	  var that = this;
		var desc = [];
		_.each(chatFlagOptions(), function(opt) {
		  if (that.get('ChatFlags')[phaseType] != null && (that.get('ChatFlags')[phaseType] & opt.id) == opt.id) {
			  desc.push(opt.name);
			}
		});
		if (phaseType != 'BeforeGame' && phaseType != 'AfterGame') {
			desc.push(_.find(deadlineOptions, function(opt) {
				return opt.value == that.get('Deadlines')[phaseType];
			}).name);
			that.appendSecrecy('SecretEmail', 'DuringGame', desc);
			that.appendSecrecy('SecretNickname', 'DuringGame', desc);
			that.appendSecrecy('SecretNation', 'DuringGame', desc);
		} else {
			that.appendSecrecy('SecretEmail', phaseType, desc);
			that.appendSecrecy('SecretNickname', phaseType, desc);
			that.appendSecrecy('SecretNation', phaseType, desc);
		}
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

