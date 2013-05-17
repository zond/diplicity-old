$(document).on( "mobileinit", function() {
	// Prevents all anchor click handling including the addition of active button state and alternate link bluring.
	$.mobile.linkBindingEnabled = false;

	// Disabling this will prevent jQuery Mobile from handling hash changes
	$.mobile.hashListeningEnabled = false;
});

