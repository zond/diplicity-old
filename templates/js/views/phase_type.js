window.PhaseTypeView = BaseView.extend({

  template: _.template($('#phase_type_underscore').html()),

	className: "panel panel-default",

	events: {
		"change .phase-deadlines": "changeDeadline",
		"hide.bs.collapse .phase": "collapse",
		"show.bs.collapse .phase": "expand",
	},

	initialize: function(options) {
		this.parentId = options.parentId;
		this.phaseType = options.phaseType;
		this.editable = options.editable;
		this.gameState = options.gameState;
		this.expanded = false;
	},

	collapse: function(ev) {
	  this.expanded = false;
	},

	expand: function(ev) {
	  this.expanded = true;
	},

	changeDeadline: function(ev) {
		this.gameState.get('Deadlines')[this.phaseType] = parseInt($(ev.target).val()); 
		this.gameState.trigger('change');
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  editable: that.editable,
			parentId: that.parentId,
		  me: that.me,
			gameState: that.gameState,
		  phaseType: that.phaseType,
			expanded: that.expanded,
		}));
		_.each(deadlineOptions, function(opt) {
		  if (that.gameState.get('Deadlines')[that.phaseType] != null && that.gameState.get('Deadlines')[that.phaseType] == opt.value) {
				that.$('.phase-deadlines').append('<option value="{0}" selected="selected">{1}</option>'.format(opt.value, opt.name));
			} else {
				that.$('.phase-deadlines').append('<option value="{0}">{1}</option>'.format(opt.value, opt.name));
			}
		});
		_.each(chatFlagOptions(), function(opt) {
			that.$('form').append(new ChatFlagView({
				gameState: that.gameState,
				phaseType: that.phaseType,
				opt: opt,
			}).doRender().el);
		});
		return that;
	},

});
