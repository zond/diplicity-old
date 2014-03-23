window.ChatMessage = Backbone.Model.extend({

  localStorage: true,

});

window.ChatMessage.channelIdFor = function(members) {
  return _.collect(members, function(x, id) {
	  return id;
	}).sort().join("-");
};
