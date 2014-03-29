window.BaseView = Backbone.View.extend({
 
	chain: [],

	fetch: function(obj) {
	  if (this.subscriptions == null) {
		  this.subscriptions = [];
		}
		this.subscriptions.push(obj);
		obj.fetch();
	},

	addChild: function(child) {
		if (this.children == null) {
			this.children = [];
		}
		this.children.push(child);
	},

	fixNavigateLinks: function() {
		this.$('a.navigate').each(function(ind, el) {
			$(el).bind('click', function(ev) {
				ev.preventDefault();
				ev.stopPropagation();
				navigate($(el).attr('href'));
			});
		});	
	},

	findParent: function() {
	  var that = this;
		that.$el.parents().each(function(x, el) {
			if (el.CurrentBaseView != null) {
			  el.CurrentBaseView.addChild(that);
			}
		});
	},

	doRender: function() {
	  var that = this;
		if (that.el != null && that.el.CurrentBaseView != null && that.el.CurrentBaseView.cid != that.cid) {
			that.el.CurrentBaseView.clean();
		}
		that.el.CurrentBaseView = that;
		that.findParent();
		that.render();
		that.fixNavigateLinks();
		if (that.rendered) {
		  that.delegateEvents();
		}
	  that.rendered = true;
		return that;
	},

	clean: function(remove) {
		if (typeof(this.onClose) == 'function') {
			this.onClose();
		}
		this.undelegateEvents();
		this.cleanChildren(remove);
		this.stopSubscribing();
		if (remove) {
		  this.remove();
		} else {
			this.stopListening();
		}
	},

	stopSubscribing: function() {
		if (this.subscriptions != null) {
			_.each(this.subscriptions, function(subs) {
				subs.close();
			});
		}
		this.children = [];
	},

	cleanChildren: function(remove) {
		if (this.children != null) {
			_.each(this.children, function(child) {
				child.clean(remove);
			});
		}
		this.children = [];
	},

});



