window.ChatMessages = Backbone.Collection.extend({

  localStorage: true,

	model: ChatMessage,

  comparator: 'CreatedAt',

});

