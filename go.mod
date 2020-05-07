module github.com/mattermost/mattermost-plugin-demo

go 1.14

require (
	4d63.com/gochecknoglobals v0.0.0-20190306162314-7c3491d2b6ec // indirect
	4d63.com/gochecknoinits v0.0.0-20200108094044-eb73b47b9fc4 // indirect
	github.com/alecthomas/gocyclo v0.0.0-20150208221726-aa8f8b160214 // indirect
	github.com/alexkohler/nakedret v1.0.0 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/cristalhq/jwt v1.1.1
	github.com/go-ldap/ldap v3.0.3+incompatible // indirect
	github.com/gogo/googleapis v1.1.0 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20200309095847-7953dde2c7bf // indirect
	github.com/jgautheron/goconst v0.0.0-20200227150835-cda7ea3bf591 // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/lyft/protoc-gen-validate v0.0.13 // indirect
	//github.com/mattermost/mattermost-server v5.9.0+incompatible
	github.com/mattermost/mattermost-plugin-api v0.0.9 // indirect
	github.com/mattermost/mattermost-server v5.11.1+incompatible
	//github.com/mattermost/mattermost-server v5.11.1+incompatible
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20200313113657-e2883bfe5f37
	github.com/mdempsky/maligned v0.0.0-20180708014732-6e39bd26a8c8 // indirect
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f // indirect
	github.com/mibk/dupl v1.0.0 // indirect
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/opennota/check v0.0.0-20180911053232-0c771f5545ff // indirect
	github.com/pkg/errors v0.9.1
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/stripe/safesql v0.2.0 // indirect
	github.com/tsenart/deadcode v0.0.0-20160724212837-210d2dc333e9 // indirect
	github.com/walle/lll v1.0.1 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	mvdan.cc/interfacer v0.0.0-20180901003855-c20040233aed // indirect
	mvdan.cc/lint v0.0.0-20170908181259-adc824a0674b // indirect
	mvdan.cc/unparam v0.0.0-20200501210554-b37ab49443f7 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
//replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
