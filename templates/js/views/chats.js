window.ChatsView = BaseView.extend({

  template: _.template($('#chats_underscore').html()),

	initialize: function(options) {
		this.rendered = false;
		this.channelViews = {};
		this.game = options.game;
		this.listenTo(this.collection, "sync", this.update);
		this.listenTo(this.collection, "reset", this.update);
		this.listenTo(this.collection, "add", this.update);
		this.listenTo(this.collection, "remove", this.update);
	},

	update: function() {
    var that = this;
		if (that.rendered) {
		  that.renderWithin(function() {
				if ((that.game.currentChatFlags() & {{.ChatFlag "Conference" }}) == {{.ChatFlag "Conference" }}) {
					var conferenceView = that.channelViews['Conference'];
					if (conferenceView == null) {
						conferenceView = new ChatChannelView({
							channel: that.game.conferenceChannel(),
							title: '{{.I "Conference" }}',
							name: 'Conference',
							collection: that.collection,
							game: that.game,
						}).doRender();
						that.channelViews['Conference'] = conferenceView;
						that.$el.append(conferenceView.el);
					}
				}
				if ((that.game.currentChatFlags() & {{.ChatFlag "Private" }}) == {{.ChatFlag "Private" }}) {
					_.each(variantNations(that.game.get('Variant')), function(nation) {
						var chatName = 'Private' + nation;
						var nationView = that.channelViews[chatName];
						if (nationView == null) {
						  var channel = {};
							channel[nation] = true;
							channel[that.game.me().Nation] = true;
							nationView = new ChatChannelView({
								channel: channel,
								title: {{.I "nations" }}[nation],
								name: chatName,
								collection: that.collection,
								game: that.game,
							}).doRender();
							that.channelViews[chatName] = nationView;
							that.$el.append(nationView.el);
						}
					});
				}
			});
		}
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		}));
		that.rendered = true;
		that.update();
		return that;
	},

});
