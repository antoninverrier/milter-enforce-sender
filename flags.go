package main

import "flag"

func setupFlags() {
	flag.StringVar(&cfg.BindProto, "proto", "tcp", "socket protocol [tcp|unix]")
	flag.StringVar(&cfg.BindHost, "host", "localhost:10025", "socket host:port or path [127.0.0.1:10025|/var/run/milter.sock]")
	flag.BoolVar(&cfg.CaseSensitiveAddresses, "csa", false, "case-sensitive address compare [true|false]")
	flag.BoolVar(&cfg.CaseSensitiveNames, "csn", false, "case-sensitive name compare [true|false]")
	flag.BoolVar(&cfg.RewriteUnauth, "rewrite-unauth", false, "rewrite mail on non-authenticated sessions [true|false]")
	flag.BoolVar(&cfg.AllowUnknown, "allow-unknown", false, "allow mails from users with no configured identity [true|false]")
	flag.StringVar(&cfg.IdentityDb, "identity-db", "/etc/milter-enforce-sender/identities", "database of allowed identities")
	flag.BoolVar(&cfg.DrynRun, "dry-run", false, "only log what to do, don't actually rewrite mails [true|false]")
	flag.StringVar(&cfg.ExtensionSperator, "separator", "+", "separator between user and mailbox in localpart")

	flag.Parse()
}
