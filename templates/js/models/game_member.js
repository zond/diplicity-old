window.GameMember = Backbone.Model.extend({
  render: function() {
	  return (
			'<select>' +
				'<option value="#">' + this.describe() + '</value>' + 
				'<option value="#">{{.I "View" }}</option>' +
        '<option value="#">{{.I "Settings" }}</option>' +
				'</option>'
		);
	},
	describe: function() {
		var phase = this.get('phase');
		var phaseInfo = '{{.I "forming"}}';
		if (phase != null) {
			phaseInfo = '{0} {1}, {2}'.format({{.I "seasons"}}[phase.season], phase.year, {{.I "phase_types"}}[phase.type]);
		}
		var nationInfo = '{{.I "undecided" }}';
		if (this.get('nation') != null) {
		  var nationInfo = {{.I "nations" }}[this.get('nation')];
		}
		return '{0}, {1}'.format(nationInfo, phaseInfo);
	},
});

