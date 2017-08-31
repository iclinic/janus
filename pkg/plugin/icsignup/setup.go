package icsignup

import (
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/plugin"
)

// Config has a URL field which stores identity verify token url
type Config struct {
	AuthURL string `json:"auth_url"`
	ApiURL string `json:"api_url"`
}

func init()  {
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
	route.AddInbound(Midleware(config.AuthURL, config.ApiURL))
	return nil
}