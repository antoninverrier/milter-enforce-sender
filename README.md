# milter-enforce-sender

This software is designed to interface with a milter-enabled MTA (Sendmail, Postfix) which receives mails from end users to prevent them from spoofing the sending identity of someone else. Instead of simply rejecting the mail, this is acheived by rewriting both enveloppe and header sender address (and display name) to one that is valid for the user.

⚠️⚠️ ***This software is alpha quality and has received very limited testing, use at your own risk !*** ⚠️⚠️


## GENERAL OPERATIONS
This milter checks if the identity of the user (username as sent by the MTA to the milter) matches the sender of the message (From: header). If not, it will be rewritten. The enveloppe sender will then be rewritten as well (or left alone if empty, ie. <>).

## HOW TO USE
 - ⚠️ Only use this milter on "submission" trafic from your own users (ports 465/587), **DO NOT use it on mail from the rest of the world** (port 25)
 - ⚠️ The order in wich you apply milters matters. As all milters that modify messages, **apply this milter BEFORE DKIM signing**, otherwise address rewritting will break DKIM signatures.
 - Logging is done on stdout, redirect it if needed.

## COMMAND-LINE OPTIONS
Pass the following option on startup to alter behaviour:
 - **proto**  **`tcp`** to listen on TCP socket, `unix` to listen on unix socket
 - **host** location of socket, **`localhost:10025`** (use something like `/var/run/milter.sock` for `unix` sockets)
 - **csa** force case normalization of email addresses (defaults to `false`), useful to enforce corporate approved capitalization (ie. the email address will be rewritten even if identical in syntax but different case, see last example in the **DATABASE FILE FORMAT** section bellow)
 - **csn** force case normalization of display names (defaults to `false`)
 - **rewrite-unauth** rewrite sender on mail from non-authenticated sessions, will use the entry with empty username from identity database (see third example in the **DATABASE FILE FORMAT** section bellow), otherwise mail from non authenticated senders is left alone (defaults to `false`)
 - **allow-unknown** whether to allow emails from users with no configured identity (defaults to `false`), if `true` such mail will not be rewritten, otherwise it will be rejected. Note that this only affects mail from authenticated users but missing from the identity database. If both `rewrite-unauth` and `allow-unknown` are true, mail from non-authenticated user will NOT be rewritten.
 - **identity-db** the location of the text file of the identity database, defaults to `/etc/milter-enforce-sender/identities` (see format in **DATABASE FILE FORMAT** section bellow), you can send SIGHUP to the milter to reload the database without restarting.
 - **dry-run** whether to actually rewriter sender in messages (defaults to `false`), if `true` the milter add debugging headers instead (`X-Milter-Enforce-Sender-Header-From` if the From: header would have been rewritten and `X-Milter-Enforce-Sender-Envelope-From` if the envelope sender would have been changed). Useful for testing the exhaustiveness of the identity database.
  - **separator** optionnal character to split the local part between the mailbox name and an extension. Extension are ignored when comparing mailbox names so that `bob+foo@example.com`and `bob+bar@example.com` are actually considered equals and no rewritting is done.


## DATABASE FILE FORMAT
The milter needs a database of sender identities. It is a plain-text file, located by default at /etc/milter-enforce-sender/identities.

An identity is a sequence of 3 lines :

```
username
email address
display name
```

Example file:
```
archer
j.archer@enterprise.starfleet
Jonathan ARCHER
picard
jl.picard@enterprise.starfleet
Jean-Luc PICARD

noreply@enterprise.starfleet
USS Enterprise
borg
@the-borg.space
The Borg
janeway
Kathryn.JANEWAY@Voyager.Starfleet
Kathryn JANEWAY
```

An empty username means this identity will be used to rewrite sender on mail from unauthenticated users (if enabled).

An @domain without localpart allow the user to send from any address in that domain.

Line separator can be \n or \r\n.

⚠️ Do not put an empty line at the end of the file.

## INSTALLATION
 - install a fairly recent version of Go (see your distro or [go.dev](https://go.dev/))
 - build using `go build`
 - install service (example Sytemd unit file `milter-enforce-sender.service` is provided)
 - configure your MTA (ex. Postfix: `smtpd_milters = inet:localhost:10025` ⚠️ beware of the order of milters)

## KNOWN BUGS & DESIGN LIMITATIONS
 - ⚠️ The milter hasn't been tested outside SMTP context (ie. by using `/usr/bin/sendmail`)
 - ⚠️ The milter doesn't enforce the sanity of values in the identities database. (no check of syntax of usernames, email addresses, etc.)
 - ⚠️ If a user has multiple identities for the same username, and they send a mail from another identity that is not referenced in the database, which identity will be selected from the database to perform the rewrite is undefined: it will be one that is valid for the user, but no way to select which one. **If the user has valid mailboxes in multiple domain, this can have effect of unexpectedly switching sender domain.** for example: Bob is allowed to use bob@example.com, bob.bobster@example.com and sales@example.net, if he tries to send a mail from alice@example.com, the sender might get rewritten to bob@example.com but it might also ends up being rewritten to sales@example.net
 - ⚠️ The display names in the database will be encoded in UTF-8 in From: header (enconding not configurable).
 - The identity database format could probably be better, however it will always, by design, be a plain text file. Querying an LDAP directory or an SQL database is beyond the scope of this milter. Use an export script instead to build the text file.

## TODO
 - [ ] ability to use LDIF file as identity database
 - [ ] ability to set a default identity for a user
 - [ ] proper build system
 - [ ] better docs

## LICENCE
milter-enforce-sender
Copyright (c) 2025 Antonin Verrier

This software is distributed under the GNU GPL version 3.0 only, without "later version" clause.
