package ads

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/mholt/caddy"
	"net"
)

func init() {
	caddy.RegisterPlugin("ads", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next()

	blocklists := make([]string, 0)
	targetIP := net.ParseIP(defaultResolutionIP)
	logBlocks := false

	for c.NextBlock() {
		value := c.Val()

		switch value {
		case "default-lists":
			blocklists = append(blocklists, defaultBlocklists...)
		case "list":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No URL found after list token"))
			}
			url := c.Val()
			blocklists = append(blocklists, url)
		case "target":
			if !c.NextArg() {
				return plugin.Error("ads", c.Err("No target IP specified"))
			}
			ip := net.ParseIP(c.Val())
			if ip == nil {
				return plugin.Error("ads", c.Err("Invalid target IP specified"))
			}
			targetIP = ip
		case "log":
			logBlocks = true
			// Do Nothing in case of { or }
		case "}":
			break
		case "{":
			break
		}
	}

	if len(blocklists) == 0 {
		blocklists = defaultBlocklists
	}

	c.OnStartup(func() error {
		once.Do(func() {
			metrics.MustRegister(c, requestCount)
			metrics.MustRegister(c, blockedRequestCount)
			metrics.MustRegister(c, requestCountBySource)
			metrics.MustRegister(c, blockedRequestCountBySource)
		})
		return nil
	})

	blockageMap, err := GenerateBlockageMap(blocklists)
	if err != nil {
		return plugin.Error("ads", c.Err("Failed to fetch blocklists"))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		adsPlugin := DNSAdBlock{
			Next:       next,
			BlockLists: blocklists,
			blockMap:   blockageMap,
			TargetIP:   targetIP,
			LogBlocks:  logBlocks,
		}

		return adsPlugin
	})

	return nil
}
