window.ChatMessage = Backbone.Model.extend({

  localStorage: true,

});

window.ChatMessage.channelIdFor = function(members) {
  return _.collect(members, function(x, nat) {
	  return nat;
	}).sort().join("-");
};
