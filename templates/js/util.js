String.prototype.format = function() {
	var args = arguments;
	return this.replace(/{(\d+)}/g, function(match, number) { 
		return typeof args[number] != 'undefined'
		? args[number]
		: match
		;
	});
};

function allocationMethods() {
	var rval = [];
	{{range .AllocationMethods}}rval.push({
		id: '{{.Id}}',
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
}

function variants() {
	var rval = [];
	{{range .Variants}}rval.push({
		id: '{{.Id}}',
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
};

function allocationMethodName(id) {
 {{range .AllocationMethods}}if (id == '{{.Id}}') {
   return '{{.Translation}}';
 }
 {{end}}
 return null;
}

function variantName(id) {
	{{range .Variants}}if (id == '{{.Id}}') {
		return '{{.Translation}}';
	}
	{{end}}
	return null;
};

function variantNations(id) {
  {{range .Variants}}if (id == '{{.Id}}') {
    return {{.JSONNations}};
	}
	{{end}}
	return null;
}

function phaseTypes(variant) {
	{{range .Variants}}if (variant == '{{.Id}}') {
		var rval = [];
		{{range .PhaseTypes}}rval.push('{{.}}');
		{{end}}
		return rval;
	}
	{{end}}
	return [];
};

function chatFlagOptions() {
	var rval = [];
	{{range .ChatFlagOptions}}rval.push({
		id: {{.Id}},
		name: '{{.Translation}}',
	});
	{{end}}
	return rval;
};

defaultAllocationMethod = '{{.DefaultAllocationMethod}}';
defaultVariant = '{{.DefaultVariant}}';
defaultDeadline = 1440;
defaultChatFlags = {{.ChatFlag "White"}} | {{.ChatFlag "Conference"}} | {{.ChatFlag "Private"}};

deadlineOptions = [
	{ value: 5, name: '{{.I "5 minutes" }}' },
	{ value: 10, name: '{{.I "10 minutes" }}' },
	{ value: 20, name: '{{.I "20 minutes" }}' },
	{ value: 30, name: '{{.I "30 minutes" }}' },
	{ value: 60, name: '{{.I "1 hour" }}' },
	{ value: 120, name: '{{.I "2 hours" }}' },
	{ value: 240, name: '{{.I "4 hours" }}' },
	{ value: 480, name: '{{.I "8 hours" }}' },
	{ value: 720, name: '{{.I "12 hours" }}' },
	{ value: 1440, name: '{{.I "24 hours" }}' },
	{ value: 2880, name: '{{.I "2 days" }}' },
	{ value: 4320, name: '{{.I "3 days" }}' },
	{ value: 5760, name: '{{.I "4 days" }}' },
	{ value: 7200, name: '{{.I "5 days" }}' },
	{ value: 10080, name: '{{.I "1 week" }}' },
	{ value: 14400, name: '{{.I "10 days" }}' },
	{ value: 20160, name: '{{.I "2 weeks" }}' },
];

function deadlineName(value) {
  var found = _.find(deadlineOptions, function(opt) {
	  return opt.value == value;
	});
	if (found != null) {
	  return found.name;
	} else {
	  return '';
	}
};

var mainButtons = [
	{
	  url: '/',
		label: '{{.I "Games" }}',
	},
	{
	  url: '/open',
		label: '{{.I "Join" }}',
	},
	{
	  url: '/create',
		label: '{{.I "Create" }}',
	},
];

function navLinks(buttons) {
  window.session.bottom_navigation.navLinks(buttons);
};

function navigate(to) {
	window.session.active_url = to;
	window.session.router.navigate(to, { trigger: true });
	window.session.bottom_navigation.update();
	$('body').css({'margin-top': (($('.navbar-fixed-top').height()) + 1 )+'px'});
	$('body').css({'margin-bottom': (($('.navbar-fixed-bottom').height()) + 1 )+'px'});
}

