window.PhaseTypeView = Backbone.View.extend({

  template: _.template($('#phase_type_underscore').html()),

  me: new Date().getTime(),

	events: {
		"change .deadline": "changeDeadline",
		"change .chat-flag": "changeChatFlag",
	},

	changeDeadline: function(ev) {
		this.game.deadlines[this.phaseType] = parseInt($(ev.target).val()); 
		if (this.gameMember != null) {
			this.gameMember.trigger('desc_change');
		}
	},

	changeChatFlag: function(ev) {
	  if ($(ev.target).is(":checked")) {
			this.game.chat_flags[this.phaseType] |= parseInt($(ev.target).attr('data-chat-flag'));
		} else {
			this.game.chat_flags[this.phaseType] = this.game.chat_flags[this.phaseType] & (~parseInt($(ev.target).attr('data-chat-flag')));
		}
		if (this.gameMember != null) {
			this.gameMember.trigger('desc_change');
		}
	},

	initialize: function(options) {
	  _.bindAll(this, 'render');
		this.phaseType = options.phaseType;
		this.game = options.game;
		this.owner = options.owner;
		this.gameMember = options.gameMember;
		if (this.gameMember != null) {
			var that = this;
			this.gameMember.bind('desc_change', function() {
				that.$('.desc').text(that.getDesc());
			});
		}
	},

	getDesc: function() {
		var desc = [];
		for (var i = 0; i < deadlineOptions.length; i++) { 
		  var opt = deadlineOptions[i];
		  if (opt.value == this.game.deadlines[this.phaseType]) {
			  desc.push(opt.name);
			}
		} 
		for (var i = 0; i < chatFlagOptions().length; i++) {
			var opt = chatFlagOptions()[i];
			if ((opt.id & this.game.chat_flags[this.phaseType]) != 0) {
			  desc.push(opt.name);
			}
		}
		return desc.join(", ");
	},

  render: function() {
		this.$el.html(this.template({
		  owner: this.owner,
		  me: this.me,
		  phaseType: this.phaseType,
			selected: this.game.deadlines[this.phaseType],
			desc: this.getDesc(),
			chatFlags: this.game.chat_flags[this.phaseType],
			deadlineOptions: deadlineOptions,
		}));
		return this;
	},

});
