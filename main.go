package main

import (
	"bufio"
	"context"
	"fmt"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v3"
	"os"
)

type OpenIDC struct {
	Issuer       string   `yaml:"issuer"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`

	oauth2   *oauth2.Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

// Initialise - makes an outbound connection to fetch the provider
// configuration from the Issuer/.well-known/configuration URL
//
// Note that the ctx is only used for the duration of this call,
// it is not stored anywhere
func (app *OpenIDC) Init(ctx context.Context) error {
	var err error

	if app.Issuer == "" {
		return fmt.Errorf("issuer is missing")
	}
	if app.ClientID == "" {
		return fmt.Errorf("client_id is missing")
	}
	if app.RedirectURL == "" {
		app.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	}
	if len(app.Scopes) == 0 {
		app.Scopes = []string{oidc.ScopeOpenID}
	}

	app.provider, err = oidc.NewProvider(ctx, app.Issuer)
	if err != nil {
		return err
	}
	app.verifier = app.provider.Verifier(&oidc.Config{ClientID: app.ClientID})
	// https://godoc.org/golang.org/x/oauth2#Config
	app.oauth2 = &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		RedirectURL:  app.RedirectURL,
		Endpoint:     app.provider.Endpoint(),
		Scopes:       app.Scopes,
	}

	return nil
}

func (app *OpenIDC) CodeToIDToken(ctx context.Context, code string) (*oidc.IDToken, error) {
	// Call out to exchange code for token
	oauth2Token, err := app.oauth2.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, err
	}

	// Parse and verify ID Token payload.
	idToken, err := app.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	return idToken, nil
}

func (app *OpenIDC) AuthCodeURL(state string) string {
	return app.oauth2.AuthCodeURL(state)
}

// Replaces yaml.UnmarshalStrict
func loadYAML(filename string, result interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true)
	return dec.Decode(result)
}

func main() {
	var app OpenIDC

	err := loadYAML("settings.yaml", &app)
	if err != nil {
		fmt.Println("Unable to load config:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := app.Init(ctx); err != nil {
		fmt.Println("Unable to initialize:", err)
		os.Exit(1)
	}
	fmt.Println("Visit this URL:", app.AuthCodeURL(""))

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter code: ")
	if !scanner.Scan() {
		fmt.Println("No input")
		os.Exit(1)
	}
	code := scanner.Text()
	if code == "" {
		fmt.Println("Aborted")
		os.Exit(1)
	}

	idToken, err := app.CodeToIDToken(ctx, code)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(2)
	}
	fmt.Println("Subject:", idToken.Subject)
	fmt.Println("Issuer:", idToken.Issuer)

	// Extract custom claims - only useful if requested in scopes
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		panic(err)
	}
	if claims.Email != "" {
		fmt.Println("Email:", claims.Email)
		fmt.Println("Email verified:", claims.Verified)
	}
}
