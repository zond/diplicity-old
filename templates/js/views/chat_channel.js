window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

	className: "panel panel-default",

	events: {
	  "click .create-message-button": "createMessage",
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

	createMessage: function(ev) {
	  var that = this;
		ev.preventDefault();
		// i have NO IDEA AT ALL why i have to use this clunky id scheme to find the body, but that.$('.new-message-body').val() never produced anything but ''
		var body = $('#new-message-' + that.name).val();
		if (body != '') {
			that.collection.create({
				Recipients: that.members,
				Body: body,
				GameId: that.model.get('Id'),
			});
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
