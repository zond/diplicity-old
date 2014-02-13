window.GameChatView = BaseView.extend({

  template: _.template($('#game_chat_underscore').html()),

	events: {
	  "click .create-channel-button": "createChannel",
	},

	initialize: function() {
	  this.listenTo(this.collection, 'add', 'addMessage');
	  this.listenTo(this.collection, 'reset', 'loadMessages');
	},

	loadMessages: function() {
	  console.log('load messages!');
	},

	addMessage: function() {
	  console.log('add message!');
	},

	createChannel: function() {
	  var members = _.filter(this.$('.new-channel-nations').val().sort(), function(val) {
		  return val != 'multiselect-all';
		});
	  this.$('#chat-channels').append(new ChatChannelView({
		  collection: this.collection,
			model: this.model,
			members: members,
		}).doRender().el);
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
