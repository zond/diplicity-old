window.GameMember = Backbone.Model.extend({
  describe: function() {
		var phase = this.get('phase');
		return '{0}, {1} {2}, {3}'.format({{.I "nations"}}[this.get('nation')], {{.I "seasons"}}[phase.season], phase.year, {{.I "phase_types"}}[phase.type]);
	},
});

