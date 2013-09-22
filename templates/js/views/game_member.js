window.GameMemberView = BaseView.extend({

  template: _.template($('#game_member_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.member = options.member;
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
			member: that.member,
		}));
		return that;
	},

});
