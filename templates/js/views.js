$(function(){
  var UserPage=Backbone.View.extend({
  el1:$(".page"),
  el2:$("#E"),
  render:function(){
    this.el1.html('hi there, the rendering worked');
	 var D=document.getElementById(el2);
	 var SVGDoc=D.getSVGDocument();
	 var SVGRoot=SVGDoc.documentElement;
	 var who=SVGRoot.firstChild.nextSibling;
	 var whoName="<"+who.nodeName;
	 var whoHow=who.attributes.item(0);
	 var whoNow=whoHow.nodeName;
	 var whoWhat=whoHow.nodeValue+">";
	 var alert(whoName+" "+whoNow+"="+whoWhat);

  }
  });
  
  var userPage=new UserPage();
  
  userPage.render();
  


});
  
