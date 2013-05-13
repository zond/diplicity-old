# godip

godip is hosted at https://github.com/zond/godip

To update the local copy (used because I don't want Google App Engine to fetch the entire test suite of 400+ megs on each deploy) I run the following command in the diplicity folder:

`rsync -av --delete --exclude 'classical/droidippy' --exclude '.git*' ~/go/src/github.com/zond/godip github.com/zond/`

