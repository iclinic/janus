package icauth

import (
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

// Config has a URL field which stores identity verify token url
type Config struct {
	URL string `json:"url"`
}

func init() {
	plugin.RegisterPlugin("auth", plugin.Plugin{
		Action: setupAuth,
	})
}

func setupAuth(route *proxy.Route, rawConfig plugin.Config) error {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return err
	}
	route.AddInbound(Midleware(config.URL))
	route.AddOutbound(OutMiddleware(config.URL))
	return nil
}
