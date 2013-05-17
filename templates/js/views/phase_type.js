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
		this.owner = options.owner;
		this.gameMember = options.gameMember;
		this.gameMember.bind('change', this.update);
	},

	changeDeadline: function(ev) {
		this.gameMember.get('game').deadlines[this.phaseType] = parseInt($(ev.target).val()); 
		this.gameMember.trigger('change');
		this.gameMember.trigger('saveme');
	},

  update: function() {
	  var that = this;
		var desc = [];
		for (var i = 0; i < deadlineOptions.length; i++) { 
		  var opt = deadlineOptions[i];
		  if (opt.value == that.gameMember.get('game').deadlines[that.phaseType]) {
			  desc.push(opt.name);
				that.$('.deadline').val('' + opt.value);
			}
		} 
		for (var i = 0; i < chatFlagOptions().length; i++) {
			var opt = chatFlagOptions()[i];
			if ((opt.id & that.gameMember.get('game').chat_flags[that.phaseType]) != 0) {
			  desc.push(opt.name);
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').attr('checked', 'checked');
			} else {
				that.$('input[type=checkbox][data-chat-flag=' + opt.id + ']').removeAttr('checked');
			}
		}
		that.$('.desc').text(desc.join(", "));
		that.$('select.deadline').val(that.gameMember.get('game').deadlines[that.phaseType]);
	},

	changeChatFlag: function(ev) {
	  if ($(ev.target).is(":checked")) {
			this.gameMember.get('game').chat_flags[this.phaseType] |= parseInt($(ev.target).attr('data-chat-flag'));
		} else {
			this.gameMember.get('game').chat_flags[this.phaseType] = this.gameMember.get('game').chat_flags[this.phaseType] & (~parseInt($(ev.target).attr('data-chat-flag')));
		}
		this.gameMember.trigger('change');
		this.gameMember.trigger('saveme');
	},

  render: function() {
		this.$el.html(this.template({
		  owner: this.owner,
		  me: this.me,
		  phaseType: this.phaseType,
		}));
		this.update();
		return this;
	},

});
