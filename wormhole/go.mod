module github.com/keys-pub/keys-ext/wormhole

go 1.13

require (
	github.com/keybase/go-keychain v0.0.0-20200325143049-65d7292bc904 // indirect
	github.com/keybase/saltpack v0.0.0-20200228190633-d75baa96bffb // indirect
	github.com/keys-pub/keys v0.0.0-20200527183651-c2daa203af2c
	github.com/keys-pub/keys-ext/http/api v0.0.0-20200527181927-f0409e2de588
	github.com/keys-pub/keys-ext/http/client v0.0.0-20200527183209-12da56e5cb21
	github.com/keys-pub/keys-ext/http/server v0.0.0-20200527183746-ca4c118983c4
	github.com/labstack/echo/v4 v4.1.16 // indirect
	github.com/pion/logging v0.2.2
	github.com/pion/sctp v1.7.6
	github.com/pion/transport v0.9.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	gortc.io/stun v1.22.1
)

// replace github.com/keys-pub/keys => ../../keys

// replace github.com/keys-pub/keys-ext/http/api => ../http/api

// replace github.com/keys-pub/keys-ext/http/client => ../http/client

// replace github.com/keys-pub/keys-ext/http/server => ../http/server
