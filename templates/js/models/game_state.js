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
		var maxMembers = variantNations(this.get('Variant')).length;
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

