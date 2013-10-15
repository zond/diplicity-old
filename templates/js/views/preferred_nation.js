window.PreferredNationView = BaseView.extend({

  template: _.template($('#preferred_nation_underscore').html()),

	tagName: 'li',

	className: 'list-group-item',

  events: {
	  "click .prioritize": "prioritize",
	},

	prioritize: function(ev) {
	  ev.preventDefault();
		this.action();
	},

	initialize: function(options) {
		this.nation = options.nation;
		this.action = options.action;
	},

  render: function() {
	  var that = this;
    that.$el.html(that.template({
		  nation: that.nation,
		}));
		return that;
	},

});
