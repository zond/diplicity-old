$(function(){
  var UserPage=Backbone.View.extend({
  el:$(".page"),
  render:function(){
    this.el.html('hi there, the rendering worked');
  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
  
  alert("loading svg");
 D=document.getElementById("E");
 SVGDoc=D.getSVGDocument();
  SVGRoot=SVGDoc.documentElement;
 who=SVGRoot.firstChild.nextSibling;
 whoName="<"+who.nodeName;
  whoHow=who.attributes.item(0);
  whoNow=whoHow.nodeName;
 whoWhat=whoHow.nodeValue+">";
  alert(whoName+" "+whoNow+"="+whoWhat);
});
  
