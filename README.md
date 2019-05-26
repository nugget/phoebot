[![CircleCI](https://circleci.com/gh/nugget/phoebot.svg?style=svg)](https://circleci.com/gh/nugget/phoebot)

## Phoebot Phoenixcraft SMP Assistant

Phoebot is an autonomous assistant that hangs out in the Phoenixcraft SMP
Minecraft [Discord server] and helps out with various tasks.  

### Functions

There are two types of commands/functions that Phoebot supports.  Some commands
are "targeted" and you need to tag `@Phoebot` in your request on a public
channel, or you need to make the request in a direct message to Phoebot
privately for it to work.  Without a tag or DM, the command will be ignored.

Other commands operate on any relevant line spoken in a public channel or in
a DM session to Phoebot.  These commands do not need to be tagged with an "@"
in order to work.

#### Version Announcement Subscriptions

Phoebot will notify individual users or channels whenever a new release of
Minecraft or PaperMC is released or made available on our [hosting provider].
You can control this behavior by subscribing or unsubscribing in a channel or
private message window.  This tagged command works like this:

```
subscribe <source> <product> [ optional recipient ]
```

So, for example, in the `#chatter` window you could say:

```
@Phoebot subscribe server.pro paper @here

@Phoebot subscribe papermc paper

@Pohebot subscribe server.pro vanilla
```

This would add a subscription to that specific channel, and whenever a new
version is detected Phoebot will send a message to the channel tagged to the
optional recipient `@here`

You can turn this off by issuing an `unsubscribe` command.

You can see what subscriptions exist for a channel by using the `list
subscriptions` command.

We will be adding more sources and products over time, including things like
popular data packs and add-ons.

#### Version Report

Asking Phoebot for a `version report` will provide a list of the most popular
versions that are being tracked by the bot.

#### Time Conversion Helper

Any time you write a date/time string in Discord, if it follows this strict
format, Phoebot will helpfully follow up with time conversions to several other
time zones which are relevant to our server users.  This makes it easier and
less error-prone to announce upcoming server events.

Just include a date-time in this form in your Discord message:

```
2019-05-25 14:45 CDT
```

The time must be in 24-hour "military" form.  Most popular time zones are
supported. If yours is not, please open an issue here on Github describing what
you need.  The time does not need to be alone on a line, it can be contained
within a longer message and the conversion will still work.

Phoebot also has support for the custom time zones "SYDNEY" and "BRISBANE" to
allow for special-handling of our favorite Australian users.

### Status Report

Asking Phoebot for a `status report` will provide details about the bot's
current code and runtime statistics.  This is for internal/developer use and
probably isn't very interesting to anyone else.

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
core Go language installed.  All dependencies are vendored and included in this
repository using go modules.  However, the preferred development environment
uses [Bazel](https://www.bazel.build) for building, testing, and deploying the
bot. All you need to set up a local development environment is to install Bazel
for your operating system.

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
