window.ChatChannelView = BaseView.extend({

  template: _.template($('#chat_channel_underscore').html()),

	className: "panel panel-default",

	initialize: function(options) {
	  this.members = options.members;
	},

  render: function() {
	  var that = this;
		var name = _.collect(that.members, function(member) {
		  return member.Id;
		}).join("-");
		var title = _.collect(that.members, function(member) {
		  return member.describe(true);
		}).join(", ");
		that.$el.html(that.template({
		  members: that.members,
			model: that.model,
			name: name,
			title: title,
		}));
		return that;
	},

});
