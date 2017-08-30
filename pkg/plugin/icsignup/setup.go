package icsignup

import (
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

var (
	adminRouter router.Router
)

func init() {
	plugin.RegisterEventHook(plugin.StartupEvent, onStartup)
	// plugin.RegisterEventHook(plugin.AdminAPIStartupEvent, onAdminAPIStartup)
	plugin.RegisterPlugin("ic_signup", plugin.Plugin{
		Action: setupICSignUp,
	})
}

func setupICSignUp(route *proxy.Route, rawConfig plugin.Config) error {
	return nil
}

// func onAdminAPIStartup(event interface{}) error {
// 	e, ok := event.(plugin.OnAdminAPIStartup)
// 	if !ok {
// 		return errors.New("Could not convert event to admin startup type")
// 	}

// 	adminRouter = e.Router
// 	return nil
// }

func onStartup(event interface{}) error {

	e, ok := event.(plugin.OnStartup)
	if !ok {
		return errors.New("Could not convert event to startup type")
	}
	
	log.Debug("Loading iClinic Signup endpoints...")

	handlers := NewHandler()
	group := e.Register.Router.Group("/v2/signup")
	{
		group.POST("/", handlers.Post())
	}
	return nil
}
