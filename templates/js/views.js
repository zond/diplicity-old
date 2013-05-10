$(function(){
  alert("blah");
  var UserPage=Backbone.View.extend({
  el:$("body"),
  render:function(){
    alert("rendered");
    this.el.html('hi there, the rendering worked');
  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
});
  
