window.DeadlineSelectView = Backbone.View.extend({

  template: _.template($('#deadline_select_underscore').html()),

	events: {
	  "change select": "changeDeadline",
	},

	initialize: function(options) {
	  _.bindAll(this, 'render');
		this.phaseType = options.phaseType;
		this.game = options.game;
	},

	changeDeadline: function(ev) {
	  this.game.deadlines[this.phaseType] = parseInt($(ev.target).val()); 
	},

  render: function() {
		this.$el.html(this.template({
		  phaseType: this.phaseType,
			selected: this.game.deadlines[this.phaseType],
			deadlineOptions: deadlineOptions,
		}));
		return this;
	},

});
