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
				navigate($(el).attr('href'));
			});
		});	
	},

	renderWithin: function(f) {
	  if (this.chain.length > 0 && this.chain[this.chain.length - 1].cid == this.cid) {
		  f();
		} else {
			this.chain.push(this);
			f();
			this.chain.pop();
		}
	},

	doRender: function() {
	  var that = this;
		that.cleanChildren();
		if (that.chain.length > 0) {
			that.chain[that.chain.length - 1].addChild(that);
		} else if (that.el != null) {
		  if (that.el.CurrentBaseView != null) {
			  if (that.el.CurrentBaseView.cid != that.cid) {
					that.el.CurrentBaseView.clean();
				}
			}
			that.el.CurrentBaseView = that;
		}
		that.renderWithin(function() {
			that.render();
		});
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



