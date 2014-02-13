window.GameChatView = BaseView.extend({

  template: _.template($('#game_chat_underscore').html()),

	events: {
	  "click .create-channel-button": "createChannel",
	},

	createChannel: function() {
	  console.log('create channel!');
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		_.each(variantNations(that.model.get('Variant')), function(nation) {
		  that.$('.new-channel-nations').append('<option value="' + nation + '">' + nation + '</option>');
		});
		that.$('.multiselect').multiselect({
		  includeSelectAllOption: true,
			onDropdownHide: function(ev) {
			  var el = $(ev.currentTarget);
			  el.css('margin-bottom', 0);
			},
			onDropdownShow: function(ev) {
			  var el = $(ev.currentTarget);
			  el.css('margin-bottom', el.find('.multiselect-container').height());
			},
		});
		return that;
	},

});
