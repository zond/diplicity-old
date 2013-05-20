diplicity
=========

Next generation Droidippy.

## Fundamental ideas

* Mobile first
* Full offline mode for reading data
 * Likely no creating of data while offline, and automatic sync when offline, but it would be nice
* One interface for iPhone, Android and web

### Goals

* Most of the features of Droidippy
* Easier adding of new maps and variants
* Full support for primarily iOS and Android
  * Via mobile web pages
  * Via native web view wrappers with push notification support
* Full functionality in regular computer browsers
* Easier operations and hosting
  * By hosting at Google App Engine
* Simpler and more maintainable code
  * By rewriting from scratch
  * By using Go instead of Java
  * Yes, this has less of a developer community, but by god the code is simpler
* Shared burden of development, maintenance and support
* Self moderation of the games
  * By using some kind of voting system in the games to silence abusive players

### Non goals

* The best computer browser experience
* Exact duplication of Droidippy features

### Anti goals

* Separate code base for each platform

## Current design

* Hosted at Google App Engine
* Backend implemented in Go
* Backend API relatively RESTful JSON/HTTP
* Frontend single page JavaScript application
* Frontend UI using jQuery Mobile widgets
* Frontend framework built using Backbone.js routes, views and models

## Running locally

Check out the code:

```
git clone git@github.com:zond/diplicity.git
```

Download [the Google App Engine SDK](https://developers.google.com/appengine/downloads) and unpack it somewhere.

Run the development app server:

```
[...]/google_appengine/dev_appserver.py .
```

Browse to http://localhost:8080/
