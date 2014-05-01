jQuery.timeago.settings.strings = {
  prefixAgo: {{.I "timeago.prefixAgo" }},
  prefixFromNow: {{.I "timeago.prefixFromNow" }},
  suffixAgo: {{.I "timeago.suffixAgo" }},
  suffixFromNow: {{.I "timeago.suffixFromNow" }},
  seconds: {{.I "timeago.seconds" }},
  minute: {{.I "timeago.minute" }},
  minutes: {{.I "timeago.minutes" }},
  hour: {{.I "timeago.hour" }},
  hours: {{.I "timeago.hours" }},
  day: {{.I "timeago.day" }},
  days: {{.I "timeago.days" }},
  month: {{.I "timeago.month" }},
  months: {{.I "timeago.months" }},
  year: {{.I "timeago.year" }},
  years: {{.I "timeago.years" }}
};

function enumerate(ary) {
  if (ary.length == 0) {
	  return "";
	} else if (ary.length == 1) {
	  return ary[0];
	} else {
	  var result = "";
	  for (var i = 0; i < ary.length; i++) {
		  result = result + ary[i];
			if (i < ary.length - 2) {
			  result = result + ", ";
			} else if (i < ary.length - 1) {
			  result = result + " {{.I "and" }} ";
			}
		}
		return result;
	}
}

String.prototype.format = function() {
	var args = arguments;
	return this.replace(/{(\d+)}/g, function(match, number) { 
		return typeof args[number] != 'undefined'
		? args[number]
		: match
		;
	});
}

function nbsp(s) {
  return s.replace(/\s/g, '&nbsp;');
}

function selEscape(sel) {
  return sel.replace("/", "\\/");
}

var allocationMethods = {{.AllocationMethods}};
var allocationMethodMap = {{.AllocationMethodMap}};
var secrecyTypesMap = {{.SecrecyTypesMap}};
var secretFlagMap = {{.SecretFlagMap}};
var chatFlagMap = {{.ChatFlagMap}};
var consequenceMap = {{.ConsequenceMap}};

function allocationMethodName(id) {
  return allocationMethodMap[id].Translation;
}

var variantMap = {{.VariantMap}};
var variants = {{.Variants}};

function consequenceOptions() {
  var rval = [];
	{{range .ConsequenceOptions}}rval.push({
	  id: {{.Id}},
		name: '{{.Translation}}',
	});
	{{end}}return rval;
};

function chatFlagOptions() {
	var rval = [];
	{{range .ChatFlagOptions}}rval.push({
		id: {{.Id}},
		name: '{{.Translation}}',
	});
	{{end}}return rval;
};

defaultAllocationMethod = '{{.DefaultAllocationMethod}}';
defaultVariant = '{{.DefaultVariant}}';
defaultDeadline = 1440;
defaultChatFlags = {{.ChatFlag "White"}} | {{.ChatFlag "Conference"}} | {{.ChatFlag "Private"}};
defaultNonCommitConsequences = {{.Consequence "NoWait"}};
defaultNMRConsequences = {{.Consequence "NoWait"}} | {{.Consequence "ReliabilityHit"}} | {{.Consequence "Surrender"}};

deadlineOptions = [
	{ value: 1, name: '{{.I "1 minute" }}' },
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
  [
		{
			label: '{{.I "Mine" }}',
			url: '/',
			activate: function() {
				return ['', 'mine/forming', 'mine/finished'].indexOf(window.session.active_url) != -1;
			},
		},
		{
			url: '/open',
			label: '{{.I "Others" }}',
			activate: function() {
				return ['open', 'closed', 'finished'].indexOf(window.session.active_url) != -1;
			},
		},
		{
			url: '/create',
			label: '{{.I "Create" }}',
		},
	],
];

function navLinks(buttons) {
  window.session.bottom_navigation.navLinks(buttons);
};

function navigate(to, notrigger) {
	window.session.active_url = /^\/?(.*)$/.exec(to)[1];
	window.session.router.navigate(to, { trigger: !!!notrigger });
	window.session.bottom_navigation.update();
}

function Poi(x, y) {
	this.x = x;
	this.y = y;
	this.add = function(p) {
		return new Poi(x + p.x, y + p.y);
	};  
	this.sub = function(p) {
		return new Poi(x - p.x, y - p.y);
	};  
	this.len = function() {
		return Math.sqrt(Math.pow(x, 2) + Math.pow(y, 2));
	};
	this.div = function(f) {
		return new Poi(x / f, y / f);
	};
	this.mul = function(f) {
		return new Poi(x * f, y * f);
	};
	this.orth = function() {
		return new Poi(-y, x);
	};
}

function Vec(p1, p2) {
	this.p1 = p1;
	this.p2 = p2;
	this.len = function() {
		return p2.sub(p1).len();
	};
	this.dir = function() {
		return p2.sub(p1).div(this.len());
	};
	this.orth = function() {
		return this.dir().orth();
	};
}
