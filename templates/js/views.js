$(function(){
  var UserPage=Backbone.View.extend({
    el1:$(".page"),
    el2:$("#map"),
    render:function(){
      this.el1.html('hi there, the rendering worked');
      alert("hi");

    }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
  });
});
  
