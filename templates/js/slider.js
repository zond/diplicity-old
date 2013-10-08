$(window).load(function() {
  $(document).on('click', '.slider .slider-control a', function(ev) {
	  ev.preventDefault();
		var link = $(ev.currentTarget);
		var content = $('#' + link.attr('data-slider-id')); 
		content.width(400);
		content.removeClass('hidden');
	});
});
