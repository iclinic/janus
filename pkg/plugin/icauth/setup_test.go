package icauth

import (
	"testing"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	route := proxy.NewRoute(&proxy.Definition{})
	err := setupAuth(route, make(plugin.Config))
	assert.NoError(t, err)

	assert.Len(t, route.Inbound, 1)
	assert.Len(t, route.Outbound, 1)
}
