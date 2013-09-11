window.GameMember = Backbone.Model.extend({

  is_member: function() {
		return window.session.user.get('Email') != '' && this.get('Email') == window.session.user.get('Email');
	},

	is_owner: function() {
		return this.get('Email') == this.get('Game').Owner;
	},

	describe: function() {
		var phase = this.get('Phase');
		var phaseInfo = '{{.I "Forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.season], phase.year, {{.I "phase_types"}}[phase.type]);
		}
		var nationInfo = '{{.I "Undecided" }}';
		if (this.get('Nation') != '') {
		  var nationInfo = {{.I "nations" }}[this.get('Nation')];
		}
		console.log('variant', this.get("Game"))
		return '{0}, {1}, {2}'.format(nationInfo, phaseInfo, variantName(this.get('Game').Variant));
	},
});

