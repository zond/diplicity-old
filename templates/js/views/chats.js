window.ChatsView = BaseView.extend({

  template: _.template($('#chats_underscore').html()),

	initialize: function(options) {
		this.rendered = false;
		this.channelViews = {};
		this.game = options.game;
		this.listenTo(window.session.user, 'change', this.doRender);
		this.listenTo(this.collection, "sync", this.update);
		this.listenTo(this.collection, "reset", this.update);
		this.listenTo(this.collection, "add", this.update);
		this.listenTo(this.collection, "remove", this.update);
		this.fetch(this.collection);
	},

	update: function() {
    var that = this;
		if (that.rendered) {
			if ((that.game.currentChatFlags() & {{.ChatFlag "Conference" }}) == {{.ChatFlag "Conference" }}) {
				var conferenceView = that.channelViews['Conference'];
				if (conferenceView == null) {
					conferenceView = new ChatChannelView({
						channel: that.game.conferenceChannel(),
						title: '{{.I "Conference" }}',
						name: 'Conference',
						collection: that.collection,
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
						nationView = new ChatChannelView({
							channel: {
								nation: true,
							},
							title: {{.I "nations" }}[nation],
							name: chatName,
							collection: that.collection,
						}).doRender();
						that.channelViews[chatName] = nationView;
						that.$el.append(nationView.el);
					}
				});
			}
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
