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

			ifaceToIps := map[string][]net.IP{}
			for _, iface := range ifaces {
				addrs, err := iface.Addrs()
				if err != nil {
					return fmt.Errorf("[%s] %w", iface.Name, err)
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
					ipv4Addrs = append(ipv4Addrs, ip)
				}

				if len(ipv4Addrs) > 0 {
					ifaceToIps[iface.Name] = ipv4Addrs
				}
			}

			const allChoice = "All"
			var ifaceSelected = allChoice
			if selectedAllNetworks == false {
				choices := lo.MapToSlice(ifaceToIps, func(iface string, ips []net.IP) choose.Choice {
					note := strings.Join(
						lo.Map(ips, func(ip net.IP, _ int) string {
							return ip.String()
						}),
						", ",
					)
					return choose.Choice{Text: iface, Note: "(" + note + ")"}
				})
				choices = append([]choose.Choice{{Text: allChoice, Note: "0.0.0.0"}}, choices...)
				ifaceSelected, err = prompt.New().
					Ask("Network").
					AdvancedChoose(choices, choose.WithHelp(true))
				if err != nil {
					return err
				}
			}

			addrIp := "0.0.0.0"
			selectedIp, found := ifaceToIps[ifaceSelected]
			if found {
				addrIp = selectedIp[0].String()
			}

			availableSubdomains := []string{}
			if ifaceSelected == allChoice {
				availableSubdomains = append(availableSubdomains, "local")
				for _, ip := range ifaceToIps {
					availableSubdomains = append(availableSubdomains, strings.ReplaceAll(ip[0].String(), ".", "-"))
				}
			} else {
				if addrIp == "127.0.0.1" || addrIp == "0.0.0.0" {
					availableSubdomains = append(availableSubdomains, "local")
				} else {
					availableSubdomains = append(availableSubdomains, strings.ReplaceAll(addrIp, ".", "-"))
				}
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
