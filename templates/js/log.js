
var logLevel = {{.LogLevel}};

var oldConsoleLog = console.log
console.log = function() {
	var ary = Array.prototype.slice.call(arguments, 0); 
	ary.unshift(new Date());
	oldConsoleLog.apply(this, ary)
}

function logFatal() {
  if (logLevel >= 0) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("FATAL");
	  console.log.apply(console, ary)
	}
}

function logError() {
  if (logLevel >= 1) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("ERROR");
	  console.log.apply(console, ary)
	}
}

function logInfo() {
  if (logLevel >= 2) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("INFO");
	  console.log.apply(console, ary)
	}
}

function logDebug() {
  if (logLevel >= 3) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("DEBUG");
	  console.log.apply(console, ary)
	}
}

function logTrace() {
  if (logLevel >= 4) {
	  var ary = Array.prototype.slice.call(arguments, 0);
		ary.unshift("TRACE");
	  console.log.apply(console, ary)
	}
}


