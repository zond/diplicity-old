window.OpenGameMembersView = BaseView.extend({

  template: _.template($('#open_game_members_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		window.session.user.bind('change', this.doRender);
		this.collection = new GameMembers([], { url: '/games/open' });
		this.collection.bind("reset", this.doRender);
		this.collection.bind("add", this.doRender);
		this.collection.bind("remove", this.doRender);
	},

	onClose: function() {
		window.session.user.unbind('change', this.doRender);
		this.collection.unbind("reset", this.doRender);
		this.collection.unbind("add", this.doRender);
		this.collection.unbind("remove", this.doRender);
	},

  render: function() {
	  var that = this;
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		that.collection.forEach(function(model) {
		  var memberView = new GameMemberView({ 
				model: model,
				button_text: '{{.I "Join" }}',
				button_action: function() {
					model.save(null, {
						success: function() {
							that.collection.remove(model);
							window.session.router.navigate('', { trigger: true });
						},
					});
				},
			}).doRender();
			memberView.$el.attr('data-role', 'collapsible');
			that.$el.append(memberView.el);
		});
		that.$el.trigger('pagecreate');
		return that;
	},

});
