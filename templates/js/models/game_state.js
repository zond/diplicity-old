window.GameState = Backbone.Model.extend({

  idAttribute: "Id",

  is_member: function() {
		return window.session.user.get('User') != '' && this.get('Member') != null && this.get('Member').User == btoa(window.session.user.get('Email'));
	},

	describe: function() {
		var phase = this.get('Phase');
		var phaseInfo = '{{.I "Forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.season], phase.year, {{.I "phase_types"}}[phase.type]);
		}
		var member = this.get('Member');
		var nationInfo = '{{.I "Undecided" }}';
		if (member != null && member.Nation != null && member.Nation != '') {
		  var nationInfo = {{.I "nations" }}[member.Nation];
		}
		return '{0}, {1}, {2}'.format(nationInfo, phaseInfo, variantName(this.get('Variant')));
	},
});

