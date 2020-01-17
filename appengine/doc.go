/*
Package main contains the code to run the server in GAE.

	##### Running locally in development mode ######
		To run locally in dev mode, in this directory (appengine) run:
			dev_appserver.py app_dev.yaml --log_level=debug --require_indexes --clear_datastore=yes
		In the clientauth directory run:
			npm start

		home gql playground: http://localhost:8080/homeschema
		member gql playground: http://localhost:8080/memberschema
		admin gql playground: http://localhost:8080/adminschema
	##### Running locally in static content mode ######
		To run locally in static dev mode, do the following:
			in the clientauth directory run: npm run build
			in the clienthome directory run: npm run build
			in the server/appengine directory run:
			dev_appserver.py app_static.yaml --log_level=debug --require_indexes --clear_datastore=yes
	##### Running in appengine ######
		to run in gcloud, do the following in the cloud shell of appengine.google.com:
			clone friendlyrervations into the cloud shell
			in the clientauth directory run: npm install; npm run build
			in the clienthome directory run: npm install; npm run build
			(one time) gcloud app deploy index.yaml
			(one time) gcloud app deploy cron.yaml
			(one time) check the gae console and wait until the indexes are built
			gcloud app deploy

	##### Setting up on an account in google cloud with email support #####
		start a g suite account with google with a single noreply@<yourdomain> account
		get a domain <yourdomain> from google for the account if you don't have it already
		transfer the domain to your new g suite account (if not already there)
		go to appengine.google.com to setup a test project
		signup for free $300 trial (if available)
		go to appengine settings for project, set max daily spending limit and authorized email as noreply@<yourdomain>
		otherwise the logs will show:
		Error sending mail: API error 3 (mail: UNAUTHORIZED_SENDER): Unauthorized sender
		add an spf record to DNS for your domain (see https://support.google.com/a/answer/33786)
			you should see "Enabled: MX Records, DKIM Signing, and SPF Validation" in google DNS settings for <yourdomain>
		deploy to the project and create a property and then create a reservation and make sure you get an email

*/
package main
