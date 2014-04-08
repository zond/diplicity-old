window.ChatMessage = Backbone.Model.extend({

  localStorage: true,

});

window.ChatMessage.channelIdFor = function(members) {
  return _.collect(members, function(x, id) {
	  return id;
	}).sort().join(".");
};

window.ChatMessage.channelTitleFor = function(game, members) {
	return _.map(members, function(x, id) {
		var memb = game.member(id);
		if (memb == null) {
			return '{{.I "Anonymous" }}';
		}
		if (game.get('State') == {{.GameState "Created"}}) {
			return memb.describe();
		} else {
			return memb.Nation;
		}
	}).sort().join(", ")
};
