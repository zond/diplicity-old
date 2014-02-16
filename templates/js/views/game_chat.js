window.GameChatView = BaseView.extend({

  template: _.template($('#game_chat_underscore').html()),

	events: {
	  "click .create-channel-button": "createChannel",
	},

	initialize: function() {
	  this.listenTo(this.collection, 'add', this.addMessage);
	  this.listenTo(this.collection, 'reset', this.loadMessages);
	},

	loadMessages: function() {
	  console.log('load messages!');
	},

	addMessage: function() {
	  console.log('add message!');
	},

	createChannel: function() {
	  var that = this;
	  var memberIds = _.filter(that.$('.new-channel-nations').val().sort(), function(val) {
		  return val != 'multiselect-all';
		});
		memberIds.push(that.model.me().Id);
		var maxMembers = variantNations(that.model.get('Variant')).length;
		if ((memberIds.length == 2 && !that.model.hasChatFlag("ChatPrivate")) ||
		    (memberIds.length == maxMembers && !that.model.hasChatFlag("ChatConference")) ||
				((memberIds.length > 2 && memberIds.length < maxMembers) && !that.model.hasChatFlag("ChatGroup")) ||
				memberIds.length < 2) {
      that.$('.create-channel-container').append('<div class="alert alert-warning fade in">' + 
			                                           '<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' + 
																								 '<strong>Hehu</strong>' + 
																								 '</div>');
		} else {
		  members = _.inject(memberIds, function(sum, id) {
			  sum[id] = true;
				return sum;
			}, {});
			that.$('#chat-channels').append(new ChatChannelView({
				collection: that.collection,
				model: that.model,
				members: members,
			}).doRender().el);
		}
	},

	disableSelector: function() {
	  var that = this;
		var sel = that.$('.new-channel-nations');
		var selectedOptions = sel.find('option:selected');
		var nonSelectedOptions = sel.find('option').filter(function() {
			return !$(this).is(':selected');
		});
		var dropdown = sel.parent().find('.multiselect-container');

		nonSelectedOptions.each(function() {
			var input = dropdown.find('input[value="' + $(this).val() + '"]');
			input.prop('disabled', true);
			input.parent('li').addClass('disabled');
		});
	},

	enableSelector: function() {
		var that = this;
		var sel = that.$('.new-channel-nations');
		var dropdown = sel.parent().find('.multiselect-container');

		sel.find('option').each(function() {
			var input = dropdown.find('input[value="' + $(this).val() + '"]');
			input.prop('disabled', false);
			input.parent('li').addClass('disabled');
		});
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		var me = that.model.me();
		if (me != null) {
			_.each(that.model.members(), function(member) {
			  if (member.Id != me.Id) {
					var opt = $('<option value="' + member.Id + '"></option>');
					opt.text(member.describe(true));
					that.$('.new-channel-nations').append(opt);
				}
			});
      var opts = {
				onDropdownHide: function(ev) {
					var el = $(ev.currentTarget);
					el.css('margin-bottom', 0);
				},
				onDropdownShow: function(ev) {
					var el = $(ev.currentTarget);
					el.css('margin-bottom', el.find('.multiselect-container').height());
				},
			};
			if ((that.model.currentChatFlags() & chatFlagMap["ChatConference"]) == chatFlagMap["ChatConference"]) {
				opts.includeSelectAllOption = true;
			}
			that.$('.new-channel-nations').multiselect(opts);
		}
		return that;
	},

});
