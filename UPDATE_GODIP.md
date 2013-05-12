# godip

godip is hosted at https://github.com/zond/godip

To update the local copy (necessary to deploy to Google App Engine) I run the following command in the diplicity folder:

`rsync -av --delete --exclude 'classical/droidippy' --exclude '.git*' ~/go/src/github.com/zond/godip github.com/zond/`

