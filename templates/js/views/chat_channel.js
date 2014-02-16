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
		that.name = _.map(that.members, function(x, id) {
		  return id;
		}).join("-");
		that.title = _.map(that.members, function(x, id) {
		  return that.model.member(id).describe(true);
		}).join(", ");
	},

  keyup: function(ev) {
		if (ev.keyCode == 13 && !ev.shiftKey && !ev.altKey) { 
		  this.createMessage(ev);
		}
	},

	createMessage: function(ev) {
	  var that = this;
		ev.preventDefault();
		// i have NO IDEA AT ALL why i have to use this clunky id scheme to find the body, but that.$('.new-message-body').val() never produced anything but ''
		var body = $('#new-message-' + that.name).val();
		if (body != '') {
		  $('#new-message-' + that.name).val('');
			that.collection.create({
				Recipients: that.members,
				Body: body,
				GameId: that.model.get('Id'),
			}, { silent: true });
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
