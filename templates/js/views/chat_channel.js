window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

	className: "panel panel-default",

	events: {
	  "click .create-message-button": "createMessage",
		"keyup .new-message-body": "keyup",
	},

	initialize: function(options) {
	  var that = this;
	  that.members = options.members;
		that.nMembers = 0;
		that.name = _.map(that.members, function(x, id) {
		  that.nMembers++;
			return id;
		}).sort().join(".");
		that.title = _.map(that.members, function(x, id) {
		  var memb = that.model.member(id);
			if (memb == null) {
			  return '{{.I "Anonymous" }}';
			}
			if (that.model.get('State') == {{.GameState "Created"}}) {
				return memb.describe();
			} else {
				return memb.Nation;
			}
		}).sort().join(", ");
	},

  keyup: function(ev) {
		if (ev.keyCode == 13 && !ev.shiftKey && !ev.altKey) { 
		  this.createMessage(ev);
		}
	},

	createMessage: function(ev) {
	  var that = this;
		ev.preventDefault();
		if (that.model.allowChatMembers(that.nMembers)) {
			// i have NO IDEA AT ALL why i have to use this clunky id scheme to find the body, but that.$('.new-message-body').val() never produced anything but ''
			var body = $('#new-message-' + that.name).val();
			if (body != '') {
				$('#new-message-' + that.name).val('');
				that.collection.create({
					RecipientIds: that.members,
					Body: body,
					GameId: that.model.get('Id'),
				}, { silent: true });
			}
		} else {
			that.$('.channel').prepend('<div class="alert alert-warning fade in">' + 
				'<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' + 
				'<strong>' +
				'{{.I "The game does not allow that particular number of members in a chat channel right now. The only types of chat allowed at the moment are {0}."}}'.format(that.model.describeCurrentChatFlagOptions()) +
				'</strong>' + 
			'</div>');
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
			model: that.model,
			name: that.name,
			title: that.title,
		}));
		return that;
	},

});
