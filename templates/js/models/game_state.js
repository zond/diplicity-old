
function variantAbbreviate(variant, cont, s) {
	var that = this;
	var mapping = {};
	switch (cont) {
		case 'nation':
			mapping = variantMap[variant].NationAbbrevs;
			break
		case 'unitType':
			mapping = variantMap[variant].UnitTypeAbbrevs;
			break
		case 'orderType':
			mapping = variantMap[variant].OrderTypeAbbrevs;
			break
	}
	var rval = _.inject(mapping, function(sum, nat, ab) {
		if (nat == s) {
			return ab;
		} else {
			return sum;
		}
	}, s);
	return rval;
}

function queryEncodePhaseState(variant, data) {
	var rval = [];
	_.each(data.Units, function(unit, prov) {
		rval.push('u' + prov + '=' + variantAbbreviate(variant, 'unitType', unit.Type) + variantAbbreviate(variant, 'nation', unit.Nation));
	});
	_.each(data.Dislodgeds, function(unit, prov) {
		rval.push('d' + prov + '=' + variantAbbreviate(variant, 'unitType', unit.Type) + variantAbbreviate(variant, 'nation', unit.Nation));
	});
	_.each(data.SupplyCenters, function(nat, prov) {
		rval.push('s' + prov + '=' + variantAbbreviate(variant, 'nation', nat));
	});
	_.each(data.Orders, function(orders, nat) {
    _.each(orders, function(order, prov) {
			rval.push('o' + prov + '=' + _.map(order, function(part) {
				return variantAbbreviate(variant, 'orderType', part);
			}).join(','));
		});
	});
	return rval.join('&');
}

window.GameState = Backbone.Model.extend({

  urlRoot: '/games',

  localStorage: function() {
	  return this.get('Id') != null;
	},

	storageFilter: function() {
	  var filtered = JSON.parse(JSON.stringify(this));
    delete(filtered.Options);
		return filtered;
	},

  queryEncode: function() {
		return queryEncodePhaseState(this.get('Variant'), this.get('Phase'));
	},

	consequences: function(typ) {
	  var cons = [];
		if ((this.get(typ + 'Consequences') & {{.Consequence "ReliabilityHit"}}) == {{.Consequence "ReliabilityHit"}}) {
		  cons.push('{{.I "Reliability hit"}}');
		}
		if ((this.get(typ + 'Consequences') & {{.Consequence "NoWait"}}) == {{.Consequence "NoWait"}}) {
		  cons.push('{{.I "No wait"}}');
		}
		if ((this.get(typ + 'Consequences') & {{.Consequence "Surrender"}}) == {{.Consequence "Surrender"}}) {
		  cons.push('{{.I "Surrender"}}');
		}
		return cons.join(", ");
	},

	me: function() {
	  if (window.session.user.get('Email') == null || window.session.user.get('Email') == "") {
		  return null;
		}
	  return _.find(this.get('Members'), function(member) {
		  return member.User.Email == window.session.user.get('Email');
		});
	},

	decorateMember: function(member) {
		member.shortDescribe = function(withNation) {
			if (withNation && member.Nation != "") {
				return member.Nation;
			}
			if (member.User.Nickname != "") {
			  return member.User.Nickname;
			}
			if (member.User.Email != "") {
			  return member.User.Email.split("@")[0];
			}
			return '{{.I "Anonymous" }}';
		};
		member.describe = function(withNation) {
			var nation = "";
			if (withNation && member.Nation != "") {
				nation = member.Nation;
			}
			var identity = "";
			if (member.User.Nickname == "" && member.User.Email == "") {
			  identity = '{{.I "Anonymous" }}';
			} else if (member.User.Nickname == "" && member.User.Email != "") {
				identity = '<' + member.User.Email + '>';
			} else if (member.User.Nickname != "" && member.User.Email == "") {
			  identity = member.User.Nickname;
			} else {
				identity = member.User.Nickname + ' <' + member.User.Email + '>';
			}
			if (nation != "" && identity != "") {
			  return nation + ' (' + identity + ')';
			} else if (nation != "") {
			  return nation;
			} else {
			  return identity;
			}
		};
		return member;
	},

	members: function() {
	  var that = this;
	  return _.map(that.get('Members'), function(member) {
		  return that.decorateMember(member);
		});
	},

	nation: function(nat) {
	  return this.decorateMember(_.find(this.get('Members'), function(member) {
		  return member.Nation == nat;
		}));
	},

	member: function(id) {
	  var m = _.find(this.get('Members'), function(member) {
		  return member.Id == id;
		});
		if (m == null) {
			return null;
		}
		return this.decorateMember(m);
	},

	memberByNation: function(nation) {
	  return this.decorateMember(_.find(this.get('Members'), function(member) {
		  return member.Nation == nation;
		}));
	},

	conferenceChannel: function() {
		var result = {};
		_.each(variantMap[this.get('Variant')].Nations, function(nation) {
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

	hasChatFlag: function(name) {
	  return (this.currentChatFlags() & chatFlagMap[name]) == chatFlagMap[name];
	},

	currentPhaseType: function() {
		if (this.get('State') != null) {
			if (this.get('State') == {{.GameState "Created" }}) {
				return 'BeforeGame';
			} else if (this.get('State') == {{.GameState "Ended" }}) {
				return 'AfterGame';
			}
			return this.get('Phase').Type;
		}
	},

	currentChatFlags: function() {
	  return this.get('ChatFlags')[this.currentPhaseType()];
	},

	describeCurrentChatFlagOptions: function() {
	  return enumerate(this.phaseTypeChatFlagOptions(this.currentPhaseType()));
	},

	phaseTypeChatFlagOptions: function(phaseType) {
	  var that = this;
		var desc = [];
		_.each(chatFlagOptions(), function(opt) {
		  if (that.get('ChatFlags')[phaseType] != null && (that.get('ChatFlags')[phaseType] & opt.id) == opt.id) {
			  desc.push(opt.name);
			}
		});
		return desc;
	},

	describePhaseType: function(phaseType) {
	  var that = this;
		var desc = that.phaseTypeChatFlagOptions(phaseType);
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

  allowChatMembers: function(n) {
		var maxMembers = variantMap[this.get('Variant')].Nations.length;
	  if (n == 2 && this.hasChatFlag("ChatPrivate")) {
			return true;
		}
		if (n == maxMembers && this.hasChatFlag("ChatConference")) {
		  return true;
		}
		if ((n > 2 && n < maxMembers) && this.hasChatFlag("ChatGroup")) {
		  return true;
		}
		return false;
	},

	describe: function() {
	  var that = this;
		var nationInfo = "";
		if (that.get('AllocationMethod') != null) {
			nationInfo = allocationMethodName(that.get('AllocationMethod'));
		}
		var member = that.me();
		if (member != null && member.Nation != null && member.Nation != '') {
		  var nationInfo = {{.I "nations" }}[member.Nation];
		}
		var phase = that.get('Phase');
		var phaseInfo = '{{.I "Forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.Season], phase.Year, {{.I "phase_types"}}[phase.Type]);
		}
		var info = [nationInfo, phaseInfo, variantMap[that.get('Variant')].Translation];
		var lastDeadline = null;
		_.each(variantMap[that.get('Variant')].PhaseTypes, function(phaseType) {
		  var thisDeadline = that.get('Deadlines')[phaseType];
			if (thisDeadline != lastDeadline) {
				info.push(deadlineName(thisDeadline));
				lastDeadline = thisDeadline;
			}
		});
		return info.join(", ");
	},

	shorten: function(part) {
		if (part == "Move") {
			return "M";
		} else if (part == "Hold") {
			return "H";
		} else if (part == "Support") {
			return "S";
		} else if (part == "Convoy") {
			return "C";
		} else if (part == "Army") {
			return "A";
		} else if (part == "Fleet") {
			return "F";
		} else {
			return part;
		}
	},

	showOrder: function(source) {
	  var that = this;

	  var unit = that.get('Phase').Units[source];
		var nation = null;
		var order = null;
		_.each(that.get('Phase').Orders, function(orders, n) {
		  _.each(orders, function(o, s) {
			  if (source == s) {
				  nation = n;
					order = o;
				}
			});
		});

    if (unit == null) {
			return nation + ': ' + source + ' ' + _.map(order, that.shorten).join(' ');
		} else {
		  return nation + ': ' + that.shorten(unit.Type) + ' ' + source + ' ' + _.map(order, that.shorten).join(' ');
		}
	},


});

