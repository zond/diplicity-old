$(function(){
  var UserPage=Backbone.View.extend({
  el1:$(".page"),
  render:function(){
    this.el1.html('hi there, the rendering worked');
  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
});
  
