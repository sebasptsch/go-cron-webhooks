# Go Cron Webhooks

All this does is store a URL and Cron Expression in a DB, restores on start and pings the URL with an empty POST request.

The cron-paired webhooks are able to be created, edited and accessed via GraphQL, this means that I don't have to worry much about the client implementation or admin UI.
