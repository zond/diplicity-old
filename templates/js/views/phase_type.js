window.PhaseTypeView = BaseView.extend({

  template: _.template($('#phase_type_underscore').html()),

  me: new Date().getTime(),

	events: {
		"change select.deadline": "changeDeadline",
		"change input.chat-flag": "changeChatFlag",
	},

	initialize: function(options) {
	  _.bindAll(this, 'doRender', 'update');
		this.phaseType = options.phaseType;
		this.editable = options.editable;
		this.gameState = options.gameState;
	},

	changeDeadline: function(ev) {
		this.gameState.get('Deadlines')[this.phaseType] = parseInt($(ev.target).val()); 
		this.update();
	},

  update: function() {
	  var that = this;
		var desc = [];
		for (var i = 0; i < deadlineOptions.length; i++) { 
		  var opt = deadlineOptions[i];
		  if (opt.value == that.gameState.get('Deadlines')[that.phaseType]) {
			  desc.push(opt.name);
				that.$('.deadline').val('' + opt.value);
			}
		} 
		for (var i = 0; i < chatFlagOptions().length; i++) {
			var opt = chatFlagOptions()[i];
			if ((opt.id & that.gameState.get('ChatFlags')[that.phaseType]) != 0) {
			  desc.push(opt.name);
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').attr('checked', 'checked');
			} else {
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').removeAttr('checked');
			}
		}
		that.$('.desc').text(desc.join(", "));
		that.$('select.deadline').val(that.gameState.get('Deadlines')[that.phaseType]);
	},

	changeChatFlag: function(ev) {
	  if ($(ev.target).is(":checked")) {
			this.gameState.get('ChatFlags')[this.phaseType] |= parseInt($(ev.target).attr('data-chat-flag'));
		} else {
			this.gameState.get('ChatFlags')[this.phaseType] = this.gameState.get('ChatFlags')[this.phaseType] & (~parseInt($(ev.target).attr('data-chat-flag')));
		}
		this.update();
	},

  render: function() {
		this.$el.html(this.template({
		  editable: this.editable,
		  me: this.me,
		  phaseType: this.phaseType,
		}));
		this.update();
		return this;
	},

});
