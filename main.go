package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/choose"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

var Log = zerolog.New(zerolog.NewConsoleWriter())

const defaultHostPort = 5050

func main() {

	cmd := &cli.Command{
		Name:      "localhost",
		Usage:     "Local https services",
		UsageText: "localhost [port]",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "host-port",
				Aliases: []string{"h"},
				Value:   defaultHostPort,
			},
			&cli.BoolFlag{
				Name:    "all-networks",
				Aliases: []string{"a"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			selectedAllNetworks := cmd.Bool("all-networks")
			hostPort := cmd.Int("host-port")
			portStr := cmd.Args().First()
			if portStr == "" {
				return fmt.Errorf("Port is required")
			}
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("Invalid port: `%s`", portStr)
			}

			ifaces, err := net.Interfaces()
			if err != nil {
				return err
			}

			_, ipnet, _ := net.ParseCIDR("198.18.0.0/15")

			ifaceToIp := map[string]net.IP{}
			for _, iface := range ifaces {
				addrs, err := iface.Addrs()
				if err != nil {
					return fmt.Errorf("[%s] %w", iface.Name, err)
				}

				// Skip docker bridges
				if strings.HasPrefix(iface.Name, "bridge") {
					continue
				}

				var ipv4Addrs []net.IP
				for _, addr := range addrs {
					ip, _, err := net.ParseCIDR(addr.String())
					if err != nil {
						continue
					}
					ip = ip.To4()
					if ip == nil {
						continue
					}
					if ip.Equal(localIp) == false && ip.IsPrivate() == false && ipnet.Contains(ip) == false {
						continue
					}
					ipv4Addrs = append(ipv4Addrs, ip)
				}

				if len(ipv4Addrs) > 0 {
					ifaceToIp[iface.Name] = ipv4Addrs[0]
				}
			}

			const allChoice = "All"
			var ifaceSelected = allChoice
			if selectedAllNetworks == false {
				choices := lo.MapToSlice(ifaceToIp, func(iface string, ip net.IP) choose.Choice {
					note := ip.String()
					return choose.Choice{Text: iface, Note: note}
				})
				choices = append([]choose.Choice{{Text: allChoice, Note: "0.0.0.0"}}, choices...)
				ifaceSelected, err = prompt.New().
					Ask("Network").
					AdvancedChoose(choices, choose.WithHelp(true))
				if err != nil {
					return err
				}
			}

			addrIp := allIp
			selectedIp, found := ifaceToIp[ifaceSelected]
			if found {
				addrIp = selectedIp
			}

			availableSubdomains := []string{}
			if ifaceSelected == allChoice {
				for _, ip := range ifaceToIp {
					availableSubdomains = append(availableSubdomains, ipToSubdomain(ip))
				}
			} else {
				availableSubdomains = append(availableSubdomains, ipToSubdomain(addrIp))
			}

			tlsCert, err := tls.X509KeyPair(cert, certKey)
			if err != nil {
				return err
			}
			return StartProxyService(ctx, tlsCert, addrIp, hostPort, port, availableSubdomains)
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		Log.Error().Msg(err.Error())
		defer os.Exit(1)
	}
}

var localIp = net.IP{127, 0, 0, 1}
var allIp = net.IP{0, 0, 0, 0}

func ipToSubdomain(ip net.IP) string {
	if ip.Equal(localIp) || ip.Equal(allIp) {
		return "local"
	}

	return strings.ReplaceAll(ip.String(), ".", "-")
}
