package main

import (
	"context"
	"log"
	"net/mail"
	"strings"

	"github.com/d--j/go-milter/mailfilter"
)

func doMilter(ctx context.Context, trx mailfilter.Trx) (decision mailfilter.Decision, err error) {

	// Get login of user
	user := trx.MailFrom().AuthenticatedUser()
	if user == "" && !cfg.RewriteUnauth {
		log.Printf("[%s] skipping unauthenticated connection", trx.Connect().Addr)
		return mailfilter.Accept, nil
	}

	// Find the user
	identities := users[user]
	final_identity := User{}
	if len(identities) == 0 {
		if cfg.AllowUnknown {
			log.Printf("[%s] no identity found for user [%s], allow as-is", trx.Connect().Addr, user)
			return mailfilter.Accept, nil
		} else {
			log.Printf("[%s] no identity found for user [%s], rejecting message", trx.Connect().Addr, user)
			if cfg.DrynRun {
				trx.Headers().Add("X-Milter-Enforce-Sender-Reject-Reason", "No identitfy for user "+user)
				return mailfilter.Accept, nil
			} else {
				return mailfilter.CustomErrorResponse(550, "5.7.1 no configured identity for sending user"), nil
			}
		}

	}

	// Get intended email sender from message header
	original_from, err := trx.Headers().Text("From")
	original_from = strings.Trim(original_from, " ")
	original_identity := User{}
	new_from := ""
	replacefrom := false

	// if From header empty, absent or unparsable, flag for replacement
	if err != nil || original_from == "" {
		log.Printf("[%s] failed to decode From header [%s] (%s) for user [%s]",
			trx.Connect().Addr, trx.Headers().Value("From"), err, user)
		replacefrom = true
	} else {
		old, err := mail.ParseAddress(original_from)
		if err != nil {
			log.Printf("[%s] failed to parse From header [%s] (%s) for user [%s]",
				trx.Connect().Addr, trx.Headers().Value("From"), err, user)
			replacefrom = true
		} else {
			original_identity.Name = old.Name
			original_identity.Mailbox, original_identity.Domain, _ = strings.Cut(old.Address, "@")

			if original_identity.Domain == "" {
				original_identity.Domain = "nodomain.invalid"
			}
			if cfg.ExtensionSperator != "" {
				original_identity.Mailbox, original_identity.Extension, _ = strings.Cut(original_identity.Mailbox, "+")
			}

		}
	}

	// Searching for sender address in identity database
	found := false
	for _, final_identity = range identities {
		if strings.EqualFold(original_identity.Domain, final_identity.Domain) &&
			(final_identity.Mailbox == "" || strings.EqualFold(original_identity.Mailbox, final_identity.Mailbox)) {
			log.Printf("[%s] found identity [%s <%s@%s>] for user [%s]",
				trx.Connect().Addr, final_identity.Name, final_identity.Mailbox, final_identity.Domain, user)
			found = true
			break
		}

	}
	// If not found, defaulting to last identity for user
	// There is always one, otherwise we would have exited earlier (either reject or no rewrite)
	if !found {
		log.Printf("[%s] no matching identity found for user [%s], defaulting to [%s <%s@%s>]",
			trx.Connect().Addr, user, final_identity.Name, final_identity.Mailbox, final_identity.Domain)
	}

	// Keep the mailbox extension
	final_identity.Extension = original_identity.Extension

	// If identity is catch-all, keep mailbox as well
	if final_identity.Mailbox == "" {
		final_identity.Mailbox = original_identity.Mailbox
	}

	// If name doesn't match, flag From header for replacement
	if !stringCaseCompare(original_identity.Name, final_identity.Name, cfg.CaseSensitiveNames) {
		log.Printf("[%s] name mismatch [%s] != [%s] for user [%s], rewriting",
			trx.Connect().Addr, final_identity.Name, original_identity.Name, user)
		replacefrom = true
	}

	// If domain doesn't match, flag From header for replacement
	if !stringCaseCompare(original_identity.Domain, final_identity.Domain, cfg.CaseSensitiveAddresses) {
		log.Printf("[%s] domain mismatch [%s] != [%s] for user [%s], rewriting",
			trx.Connect().Addr, final_identity.Domain, original_identity.Domain, user)
		replacefrom = true
	}

	// We compare original mailbox name only if supplied (the identity database can have @domain.tld for catch-all)
	if final_identity.Mailbox != "" && !stringCaseCompare(original_identity.Mailbox, final_identity.Mailbox, cfg.CaseSensitiveAddresses) {
		log.Printf("[%s] mailbox mismatch [%s] != [%s] for user [%s], rewriting",
			trx.Connect().Addr, final_identity.Mailbox, original_identity.Mailbox, user)
		replacefrom = true
	}

	// Do actual From header replacement
	if replacefrom {
		new_from = final_identity.fromHeader()
		log.Printf("[%s] rewriting From header [%s] => [%s] for user [%s]",
			trx.Connect().Addr, original_from, new_from, user)

		if cfg.DrynRun {
			trx.Headers().Add("X-Milter-Enforce-Sender-Header-From", new_from)

		} else {
			trx.Headers().Set("From", new_from)
		}
	} else {
		log.Printf("[%s] allowing From header [%s] for user [%s]",
			trx.Connect().Addr, original_from, user)
	}

	// Do we need to change the envelope from as well?
	old_smtp_from := trx.MailFrom().Addr
	if old_smtp_from == "" {
		log.Printf("[%s] skipping empty envelope from for user [%s]", trx.Connect().Addr, user)
	} else {
		if !stringCaseCompare(old_smtp_from, final_identity.smtpFrom(), cfg.CaseSensitiveAddresses) {
			log.Printf("[%s] rewriting envelope [%s] => [%s] for user [%s]",
				trx.Connect().Addr, old_smtp_from, final_identity.smtpFrom(), user)
			if cfg.DrynRun {
				trx.Headers().Add("X-Milter-Enforce-Sender-Envelope-From", final_identity.smtpFrom())
			} else {
				trx.ChangeMailFrom(final_identity.smtpFrom(), "")
			}
		}
	}

	return mailfilter.Accept, nil
}
