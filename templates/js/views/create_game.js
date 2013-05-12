window.CreateGameView = Backbone.View.extend({

  template: _.template($('#create_game_underscore').html()),

  phaseTypes: function(variant) {
	  {{range .Variants}}if (variant == '{{.Id}}') {
		  var rval = [];
			{{range .PhaseTypes}}rval.push('{{.}}');
			{{end}}
			return rval;
		}
		{{end}}
		return [];
	},

  events: {
	  "click .create-game-button": "createGame",
	},

	createGame: function(ev) {
		var that = this;
	  this.collection.create({
		  game: {
				variant: this.$('select.create-game-variant').val(),
				private: this.$('select.create-game-private').val() == 'true',
			}
		});
		$.mobile.changePage('#home');
	},

	initialize: function(options) {
	  _.bindAll(this, 'render');
	},

  render: function() {
		this.$el.html(this.template({}));
		_.each(this.phaseTypes(this.$('.create-game-variant').val()), function(type) {
		  this.$('.deadlines').append(new DeadlineSliderView({ phaseType: type }).render().el);
		});
		this.$('.deadlines').trigger('create');
		this.delegateEvents();
		return this;
	},

});
