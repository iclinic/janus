package icsignup

import (
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

// Config has a URL field which stores identity verify token url
type Config struct {
	CreateUserURL   string `json:"createuser_url"`
	DeleteUserURL   string `json:"deleteuser_url"`
	SubscriptionURL string `json:"subscription_url"`
}

func init() {
	plugin.RegisterPlugin("signup", plugin.Plugin{
		Action: setupSignup,
	})
}

func setupSignup(route *proxy.Route, rawConfig plugin.Config) error {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return err
	}
	route.AddInbound(Midleware(config.CreateUserURL, config.DeleteUserURL, config.SubscriptionURL))
	return nil
}
