$(function(){
	
  var User=Backbone.Model.extend({
  	defaults: {
  		name:'Jospeh Schmoseph',
  		userId:'DipPlayer223',
  	}
  });	
	
  var ItemListView=Backbone.View.extend({
	  el1:$("#itemListView"),
	  render:function(){
	    this.el1.html(user.get("userId"));
	  }
  });
  var user=new User();
  var itemListView=new ItemListView();
  itemListView.render();



});
  
