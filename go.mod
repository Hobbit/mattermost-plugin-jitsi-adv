module github.com/mattermost/mattermost-plugin-demo

go 1.14

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/cristalhq/jwt v1.1.1
	github.com/go-ldap/ldap v3.0.3+incompatible // indirect
	github.com/gogo/googleapis v1.1.0 // indirect
	github.com/lyft/protoc-gen-validate v0.0.13 // indirect
	//github.com/mattermost/mattermost-server v5.9.0+incompatible
	github.com/mattermost/mattermost-plugin-api v0.0.9 // indirect
	github.com/mattermost/mattermost-server v5.11.1+incompatible
	//github.com/mattermost/mattermost-server v5.11.1+incompatible
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20200313113657-e2883bfe5f37
	github.com/nicksnyder/go-i18n v1.10.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
//replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
