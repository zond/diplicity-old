window.GameOrdersView = BaseView.extend({

  template: _.template($('#game_orders_underscore').html()),

	shorten: function(part) {
		if (part == "Move") {
			return "M";
		} else if (part == "Hold") {
			return "H";
		} else if (part == "Support") {
			return "S";
		} else if (part == "Convoy") {
			return "C";
		} else if (part == "Army") {
			return "A";
		} else if (part == "Fleet") {
			return "F";
		} else {
			return part;
		}
	},

	showOrder: function(nation, source) {
	  var that = this;

	  var unit = that.model.get('Phase').Units[source];
	  var order = _.collect(that.model.get('Phase').Orders[nation][source], that.shorten);

    if (unit == null) {
			return nation + ': ' + source + ' ' + order.join(' ');
		} else {
		  return nation + ': ' + that.shorten(unit.Type) + ' ' + source + ' ' + order.join(' ');
		}
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		}));
		console.log(that.model.get('Phase'));
		var me = that.model.me();
		if (me != null) {
			_.each(that.model.get('Phase').Orders, function(orders, nation) {
			  _.each(orders, function(order, source) {
					that.$('.orders').append(that.showOrder(nation, source, order) + '\n');
				});
			});
		}
		return that;
	},

});
