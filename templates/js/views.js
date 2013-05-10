$(function(){
  var UserPage=Backbone.View.extend({
  el:'.page',
  render:function(){
    this.$el.html('hi there, the rendering worked');
  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
});
  
