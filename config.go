package main

type MilterConfig struct {
	CaseSensitiveAddresses bool
	CaseSensitiveNames     bool
	BindProto              string
	BindHost               string
	RewriteUnauth          bool
	AllowUnknown           bool
	IdentityDb             string
	DrynRun                bool
	ExtensionSperator      string
}
