window.CreateGameView = Backbone.View.extend({

  template: _.template($('#create_game_underscore').html()),

  variants: function() {
	  var rval = [];
	  {{range .Variants}}rval.push({
		  id: '{{.Id}}',
			name: '{{.Translation}}',
		});
		{{end}}
		return rval;
	},

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
		_.each(this.variants(), function(variant) {
			this.$('select.create-game-variant').append('<option value="{0}">{{.I "Variant"}}: {1}</option>'.format(variant.id, variant.name));
		});
		_.each(this.phaseTypes(this.$('.create-game-variant').val()), function(type) {
		  this.$('.deadlines').append(new DeadlineSelectView({ phaseType: type }).render().el);
		});
		this.$el.trigger('create');
		this.delegateEvents();
		return this;
	},

});
