window.OpenGameStatesView = BaseView.extend({

  template: _.template($('#open_game_states_underscore').html()),

	initialize: function(options) {
	  _.bindAll(this, 'doRender');
		this.listenTo(window.session.user, 'change', this.doRender);
		this.collection = new GameStates([], { url: '/games/open' });
		this.listenTo(this.collection, "reset", this.doRender);
		this.listenTo(this.collection, "add", this.doRender);
		this.listenTo(this.collection, "remove", this.doRender);
		this.fetch(this.collection);
	},

  render: function() {
	  var that = this;
		navLinks(mainButtons);
		that.$el.html(that.template({
		  user: window.session.user,
		}));
		that.collection.forEach(function(model) {
		  var save_call = function() {
				model.save(null, {
					success: function() {
						navigate('/');
					},
				});
			};
		  var stateView = new GameStateView({ 
				model: model,
				parentId: 'open_games',
				editable: false,
				button_text: '{{.I "Join" }}',
				button_action: function() {
				  model.set('Members', [
						{
							UserId: btoa(window.session.user.get('Email')),
							User: {},
						}
					]);
					if (model.get('AllocationMethod') == 'preferences') {
            new PreferencesAllocationDialogView({ 
							gameState: model,
							done: function(nations) {
								model.get('Members')[0].PreferredNations = nations;
                save_call();
							},
						}).display();
					} else {
					  save_call();
					}
				},
			}).doRender();
			that.$('#open_games').append(stateView.el);
		});
		if (window.session.user.loggedIn() && that.collection.length == 0) {
		  that.$el.append('<a href="/create" class="btn btn-primary btn-lg btn-block">{{.I "No open games, why not create one?" }}</a>');
		}
		return that;
	},

});
