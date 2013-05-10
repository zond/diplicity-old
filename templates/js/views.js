$(function(){
  var UserPage=Backbone.View.extend({
  el1:$(".page"),
  el2:$("#map"),
  render:function(){
    this.el1.html('hi there, the rendering worked');
    this.el2.html('and this is a map');
  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
});
  
