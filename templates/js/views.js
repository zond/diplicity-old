$(function(){
  var UserPage=Backbone.View.extend({
    el1:$(".page"),
    el2:$("#map"),
    render:function(){
      this.el1.html('hi there, the rendering worked');
      $(document).ready ( function () {
        this.el2.svg({loadURL: 'https://github.com/cmconnor/diplicity/blob/master/img/maps/standard.svg'});
      });
    }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
  });
});
  
