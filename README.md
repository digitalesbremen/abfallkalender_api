# Bremer Abfallkalender API

[![Docker build and push](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker-build-and-push.yml/badge.svg)](https://github.com/digitalesbremen/abfallkalender_api/actions/workflows/docker-build-and-push.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This project is still alpha an in development. 

![Idea](misc/images/idea.png)

```shell
# Login to heroku
$ heroku container:login

# Create app (if not existing)
$ heroku create bremer-abfallkalender-api

# Push docker container to heroku registry
$ heroku container:push web --app bremer-abfallkalender-api

# Release app
$ heroku container:release web

# Open app
$ heroku open --app bremer-abfallkalender-api

# Logs
$ heroku logs -n 200 --app bremer-abfallkalender-api
$ heroku logs --tail --app bremer-abfallkalender-api
```