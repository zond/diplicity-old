window.GameResultsView = BaseView.extend({

  template: _.template($('#game_results_underscore').html()),

	initialize: function() {
	  this.reloadModel(this.model);
	},

	reloadModel: function(model) {
	  this.stopListening();
		this.model = model;
		this.listenTo(this.model, 'change', this.doRender);
		this.listenTo(this.model, 'reset', this.doRender);
		this.doRender();
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		if (that.model.get('Phase') != null) {
		  var lines = [];
		  var ordersByProvince = {};
			_.each(that.model.get('Phase').Orders, function(orders, nation) {
        _.each(orders, function(order, source) {
				  ordersByProvince[source] = order;
				});
			});
			_.each(that.model.get('Phase').Resolutions, function(result, source) {
			  lines.push(that.model.showOrder(source) + ': ' + result);
			});
			lines.sort();
			that.$('.results').text(lines.join('\n'));
		}
		return that;
	},

});
