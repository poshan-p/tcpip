package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/urfave/cli"
)

func Init() {

	app := &cli.App{
		Name:                 "TCP/IP",
		Usage:                "Implementation of the TCP/IP stack",
		CommandNotFound:      HandleCommandNotFound,
		EnableBashCompletion: true,
		Commands: []cli.Command{
			{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Show help message",
				Action:  ShowHelp,
			},
			{
				Name:  "show",
				Usage: "Show information",
				Subcommands: []cli.Command{
					{
						Name:   "topology",
						Usage:  "Show current topology",
						Action: ShowTopology,
					},
					{
						Name:   "node",
						Usage:  "Show node information",
						Action: ShowNode,
						Subcommands: []cli.Command{
							{
								Name:   "arp",
								Usage:  "Show ARP table of the node",
								Action: ShowNodeArpTable,
							},
							{
								Name:   "mac",
								Usage:  "Show MAC table of the node",
								Action: ShowNodeMacTable,
							},
							{
								Name:   "routing-table",
								Usage:  "Show routing table of the node",
								Action: ShowNodeRoutingTable,
							},
						},
					},
				},
			},
			{
				Name:  "run",
				Usage: "Run a command",
				Subcommands: []cli.Command{
					{
						Name:  "node",
						Usage: "Run commands on a node",
						Subcommands: []cli.Command{
							{
								Name:   "resolve-arp",
								Usage:  "Resolve ARP for a node",
								Action: RunResolveARPCommand,
							},
							{
								Name:   "ping",
								Usage:  "Ping a node",
								Action: RunPingCommand,
								Subcommands: []cli.Command{
									{
										Name:   "tunnel",
										Usage:  "Ping a node using tunneling",
										Action: RunPingTunnelCommand,
									},
								},
							},
						},
					},
				},
			},
			{
				Name:  "config",
				Usage: "Configure the network",
				Subcommands: []cli.Command{
					{
						Name:  "node",
						Usage: "Configure a node",
						Subcommands: []cli.Command{
							{
								Name:   "route",
								Usage:  "Add a route to a node",
								Action: ConfigNodeRoute,
							},
						},
					},
				},
			},
			{
				Name:   "exit",
				Usage:  "Exit application",
				Action: ExitCommand,
			},
		},
	}

	// Set up signal handling for graceful termination
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a readline instance for user input
	rl, err := readline.New("tcpip> ")
	if err != nil {
		fmt.Println("Error creating readline instance:", err)
		os.Exit(1)
	}
	defer func(rl *readline.Instance) {
		err := rl.Close()
		if err != nil {
			fmt.Println("Error closing readline instance:", err)
		}
	}(rl)

	// Run the CLI application in a goroutine
	go func() {
		err := app.Run(os.Args)
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Application loop for handling user input
applicationLoop:
	for {
		select {
		case <-signalChan:
			fmt.Println("Exiting the application")
			break applicationLoop
		default:
			line, err := rl.Readline()
			if errors.Is(err, readline.ErrInterrupt) {
				fmt.Println("Interrupted. Type 'exit' to exit.")
			} else if err != nil {
				fmt.Println("Error reading input:", err)
			}

			line = strings.TrimSpace(line)

			if line != "" {
				args := strings.Fields(line)

				err := app.Run(append([]string{os.Args[0]}, args...))
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

}
