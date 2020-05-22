# OpenIDC CLI demonstrator

This is a demonstrator of using OpenIDC to verify identity, i.e. as a
Relying Party (RP), using just the command line.

Configure `settings.yaml`: at minimum it needs the `issuer`, `client_id`,
`client_secret`.

Run the tool:

```
./openidc-cli
```

It will show a URL for you to put in your browser.  Once you have identified
yourself to your identity provider, it will give you a sign-in code.

Paste this code back into the CLI.  It will then exchange this for an ID
token at your identity provider, and show the results.  (This is the only
part which uses the client secret)

## Configuration: Google

OAuth client IDs are created at
<https://console.developers.google.com/apis/credentials/oauthclient>

When the identity provider is asked to provide a pasteable code, the
redirect_url is set to the well-known value `urn:ietf:wg:oauth:2.0:oob`

For this to work with Google, you should create your OAuth2 client as
type "Desktop App" rather than "Web application".

## Configuration: Dex

You can allow the well-known URN explicitly:

```
- id: openidc-cli
  redirectURIs:
  - urn:ietf:wg:oauth:2.0:oob
  name: 'OpenIDC CLI'
  secret: ZXhhbXBsZS1hcHAtc2VjcmV0
```

Or you can set `public: true` which permits the well-known value plus any
`http://localhost` URL.
