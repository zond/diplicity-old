window.ChatChannelView = BaseView.extend({

  className: 'btn btn-default btn-block',

	tagName: 'button',

  template: _.template($('#chat_channel_underscore').html()),

	initialize: function(options) {
		this.listenTo(window.session.user, 'change', this.doRender);
		this.listenTo(this.collection, "sync", this.doRender);
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
	  this.channel = options.channel;
		this.title = options.title;
		this.name = options.name;
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  title: that.title,
			name: that.name,
		}));
		return that;
	},

});
