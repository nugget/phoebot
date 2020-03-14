[![CircleCI](https://circleci.com/gh/nugget/phoebot.svg?style=svg)](https://circleci.com/gh/nugget/phoebot)

## Phoebot Phoenixcraft SMP Assistant

Phoebot is an autonomous assistant that hangs out in the Phoenixcraft SMP
Minecraft [Discord server] and helps out with various tasks.

For user documentaton and more information, please see [The
Wiki](https://github.com/nugget/phoebot/wiki).

### Installation and Operation

* To add the bot to a Discord server, visit [this activation link].

Phoebot is distributed as a docker image hosted on [Docker Hub].

It's running in a Kubernetes Cluster using the object definitions found in the
`k8s` directory.  Kubernetes is not required for operation, though.  It can be
run in any Docker or Docker-complaint container envrionment.  During runtime,
the bot will make use of the following environment variables:

* `DISCORD_BOT_TOKEN` is your authentication token for the application/bot that
  you create on the [Discord developer portal].
 
* `MC_CHECK_INTERVAL` (optional) controls how frequently Phoebot will check for
  updated versions of packages.

* `PHOEBOT_DEBUG` (optional) causes the bot to start up with debug log level
  instead of waiting for an operator to issue that command in chat.  See
  "Console Logging" below.

* `STATE_FILENAME` (defaults to `/phoebot/phoebot-state.xml`) allows you to use
  a different location for the state file for easier local development.

## Developer Notes

### Building Phoebot

Phoebot is written in [Go](https://golang.org) and can be built with just the
core Go language installed.  The preferred development environment uses
[Bazel](https://www.bazel.build) for building, testing, and deploying the bot.
All you need to set up a local development environment is to install Bazel for
your operating system.

The root level `Makefile` contains targets for common build operations.

`make run` will build and run the bot locally for testing.

`make deploy` is what I use to deploy the docker image to docker hub and update
Kubernetes to run the new code.  This will only work in my production
environment.  It's not set up to be generally useful.

### Console Logging

You can send the command `set loglevel to <level>` on Discord to dynamically
change the verbosity of the console logging.  This is useful for debugging.

[this activation link]: https://discordapp.com/oauth2/authorize?client_id=581247665933779013&scope=bot&permissions=150528
[Docker Hub]: https://cloud.docker.com/u/nugget/repository/docker/nugget/phoebot
[Application Form]: https://docs.google.com/forms/d/e/1FAIpQLSdvj5J4vLsOIuvWof3B4jiZYXXpFKfsZMMSUtwKjTN5ThXDRw/viewform
[Discord server]: https://discord.gg/a6KnJcj
[hosting provider]: https://server.pro/
[Discord developer portal]: https://discordapp.com/developers/applications



/data modify block x y z CustomName set value '{"text":"MOO"}'

