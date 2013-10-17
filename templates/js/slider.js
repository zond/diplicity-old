$(window).load(function() {
  $(window).on('orientationchange', function(ev) {
	  $('.slider-content').addClass('hidden');
	});
  $(document).on('click', '.slider-controls a', function(ev) {
	  ev.preventDefault();
		var singleMode = false;
	  var windowWidth = $(window).width();
		if (windowWidth < 1240) {
		  singleMode = true;
		}
		var link = $(ev.currentTarget);
		var content = $('#' + link.attr('data-slider-id')); 
		if (content.hasClass('hidden')) {
		  if (singleMode) {
			  $('.slider-content').addClass('hidden');
			}
			var width = $(window).width() - 40;
			if (width > 400) {
			  width = 400;
			}
			content.width(width);
			content.removeClass('hidden');
		} else {
		  content.addClass('hidden');
		}
	});
});
