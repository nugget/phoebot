[![CircleCI](https://circleci.com/gh/nugget/phoebot.svg?style=svg&circle-token=672825c48e9ccd415262e2777de633518e4543bd)](https://circleci.com/gh/nugget/phoebot)

## Phoebot Phoenixcraft SMP Assistant

Phoebot is an autonomous assistant that hangs out in the Phoenixcraft SMP
Minecraft [Discord server] and helps out with various tasks.  

### Installation

To add the bot to a Discord server, visit [this activation link].

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
```

This would add a subscription to that specific channel, and whenever a new
version is detected Phoebot will send a message to the channel tagged to the
optional recipient `@here`

You can turn this off by issuing an `unsubscribe` command.

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

## Developer Notes

### Console Logging

You can send the command `set loglevel to <level>` on Discord to dynamically
change the verbosity of the console logging.  This is useful for debugging.

[this activation link]: https://discordapp.com/oauth2/authorize?client_id=581247665933779013&scope=bot&permissions=150528
[Application Form]: https://docs.google.com/forms/d/e/1FAIpQLSdvj5J4vLsOIuvWof3B4jiZYXXpFKfsZMMSUtwKjTN5ThXDRw/viewform
[Discord server]: https://discord.gg/a6KnJcj
[hosting provider]: https://server.pro/
