$(window).load(function() {
  $(document).on('click', '.slider-controls a', function(ev) {
	  ev.preventDefault();
		var link = $(ev.currentTarget);
		var content = $('#' + link.attr('data-slider-id')); 
		if (content.hasClass('hidden')) {
			content.width(400);
			content.removeClass('hidden');
		} else {
		  content.addClass('hidden');
		}
	});
});
