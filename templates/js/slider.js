$(window).load(function() {
  var normWidth = 500;
	var buttonWidth = 50;
  var resizeSliders = function() {
    var win = $(window).width();
	  var count = $('.slider-content.visible').length;
		while (win < normWidth * count + buttonWidth && count > 1) {
			var eldest = null;
			var eldestAt = new Date().getTime();
			$('.slider-content.visible').each(function(ind, el) {
				if (eldestAt > parseInt($(el).attr('data-displayed-at'))) {
					eldestAt = parseInt($(el).attr('data-displayed-at'));
					eldest = $(el);
				}
			});
			console.log('hiding one');
			eldest.removeClass('visible').addClass('hidden');
			count = $('.slider-content.visible').length;
		}
		if (count == 1 && win < normWidth + buttonWidth) {
			$('.slider-content.visible').width(win - buttonWidth);
		} else {
			$('.slider-content.visible').width(normWidth);
		}
	};
  $(window).on('orientationchange', resizeSliders);
	$(window).on('resize', resizeSliders);
  $(document).on('click', '.slider-controls a', function(ev) {
	  ev.preventDefault();
		var link = $(ev.currentTarget);
		var content = $('#' + link.attr('data-slider-id')); 
		if (content.hasClass('hidden')) {
			content.removeClass('hidden').addClass('visible');
			content.attr('data-displayed-at', new Date().getTime());
		} else {
		  content.addClass('hidden').removeClass('visible');
		}
		resizeSliders();
	});
});
