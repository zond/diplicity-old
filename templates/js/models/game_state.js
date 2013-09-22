window.GameState = Backbone.Model.extend({

  idAttribute: "Id",

	me: function() {
	  if (window.session.user.get('Email') == null) {
		  return null;
		}
	  return _.find(this.get('Members'), function(member) {
		  return member.UserId == btoa(window.session.user.get('Email'));
		});
	},

	describe: function() {
		var phase = this.get('Phase');
		var phaseInfo = '{{.I "Forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.season], phase.year, {{.I "phase_types"}}[phase.type]);
		}
		var member = this.get('Member');
		var nationInfo = allocationMethodName(this.get('AllocationMethod'));
		if (member != null && member.Nation != null && member.Nation != '') {
		  var nationInfo = {{.I "nations" }}[member.Nation];
		}
		return '{0}, {1}, {2}'.format(nationInfo, phaseInfo, variantName(this.get('Variant')));
	},
});

