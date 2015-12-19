package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vikstrous/zengge-lightcontrol/control"
	"github.com/vikstrous/zengge-lightcontrol/local"
	"github.com/vikstrous/zengge-lightcontrol/manage"
	"github.com/vikstrous/zengge-lightcontrol/remote"
)

func addAll(parent *cobra.Command, children []cobra.Command) {
	for _, child := range children {
		child := child
		parent.AddCommand(&child)
	}
}

func main() {

	var mac string
	var devID string
	var host string

	zenggeCmd := &cobra.Command{
		Use:   "zengge-lightcontrol",
		Short: "zengge-lightcontrol is a command line tool for controlling zengge lights",
		Long:  `zengge-lightcontrol is a command line tool for controlling zengge lights`,
		//Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		//},
	}

	var rc *remote.Controller
	var controller *control.Controller
	var manager *manage.Manager

	localCmd := &cobra.Command{
		Use:   "local",
		Short: "local commands",
		Long:  `Execute commands over the local network`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if host == "" {
				return fmt.Errorf("Auto-discovery not implemented yet. Please specify a host with --host")
			}
			transport, err := local.NewTransport(host)
			if err != nil {
				return fmt.Errorf("Failed to connect. %s", err)
			}
			controller = &control.Controller{transport}
			return nil
		},
	}

	manageCmd := &cobra.Command{
		Use:   "manage",
		Short: "management commands",
		Long:  `Execute management commands over the local network`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if host == "" {
				return fmt.Errorf("Auto-discovery not implemented yet. Please specify a host with --host")
			}
			var err error
			manager, err = manage.NewManager(host)
			return err
		},
	}

	manageCmds := []cobra.Command{
		{
			Use:   "help",
			Short: "print the router's help menu",
			RunE: func(cmd *cobra.Command, args []string) error {
				help, err := manager.Help()
				if err != nil {
					return err
				}
				fmt.Println(help)
				return nil
			},
		},
		{
			Use:   "shell",
			Short: "interactive shell to the router",
			RunE: func(cmd *cobra.Command, args []string) error {
				return manager.Shell()
			},
		},
		{
			Use:   "get-wsinfo",
			Short: "wifi station SSID and password",
			RunE: func(cmd *cobra.Command, args []string) error {
				ssid, password, err := manager.GetWSInfo()
				if err != nil {
					return err
				}
				fmt.Printf("SSID: %s\n", ssid)
				fmt.Printf("Pass: %s\n", password)
				return nil
			},
		},
		{
			Use:   "http-send",
			Short: "send an http request; sockB must be closed for this to work",
			RunE: func(cmd *cobra.Command, args []string) error {
				response, err := manager.HTTPSend("http://93.184.216.34", "80", "GET", "/", "keep-alive", "derp", "")
				if err != nil {
					return err
				}
				fmt.Print(response)
				return nil
			},
		},
	}

	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "remote commands",
		Long:  `Execute commands through the remote control service`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if devID == "" {
				return fmt.Errorf("devid required")
			}
			rc = remote.NewController("http://wifi.magichue.net/WebMagicHome/ZenggeCloud/ZJ002.ashx", "8ff3e30e071c9ef5b304d83239d0c707", devID)
			rc.Login()
			return nil
		},
	}

	remoteCmds := []cobra.Command{
		{
			Use:   "get-devices",
			Short: "get the list of devices associated with your account",
			RunE: func(cmd *cobra.Command, args []string) error {
				devices, err := rc.GetDevices()
				if err != nil {
					return err
				}
				fmt.Println(devices)
				return nil
			},
		},
		{
			Use:   "register",
			Short: "register a new device",
			RunE: func(cmd *cobra.Command, args []string) error {
				if mac == "" {
					return fmt.Errorf("mac required")
				}
				return rc.RegisterDevice(mac)
			},
		},
		{
			Use:   "deregister",
			Short: "deregister a device",
			RunE: func(cmd *cobra.Command, args []string) error {
				if mac == "" {
					return fmt.Errorf("mac required")
				}
				return rc.DeregisterDevice(mac)
			},
		},
		{
			Use:   "get-owners",
			Short: "list everyone able to control a given device",
			RunE: func(cmd *cobra.Command, args []string) error {
				if mac == "" {
					return fmt.Errorf("mac required")
				}
				owners, err := rc.GetOwners(mac)
				if err != nil {
					return err
				}
				fmt.Println(owners)
				return nil
			},
		},
	}

	localCmd.PersistentFlags().StringVarP(&host, "host", "o", "", "ip and port of lightbulb")
	manageCmd.PersistentFlags().StringVarP(&host, "host", "o", "", "ip and port of lightbulb")
	remoteCmd.PersistentFlags().StringVarP(&mac, "mac", "m", "", "mac address of lightbulb")
	remoteCmd.PersistentFlags().StringVarP(&devID, "devid", "d", "", "devID to log in as")

	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "execute a lightbulb command remotely",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := remoteCmd.PersistentPreRunE(cmd, args)
			if err != nil {
				return err
			}
			if mac == "" {
				return fmt.Errorf("mac required")
			}
			remote_ := remote.NewRemoteTransport(rc, mac)
			controller = &control.Controller{remote_}
			return nil
		},
	}

	execCmds := []cobra.Command{
		{
			Use:   "set-color",
			Short: "change the color of the lightbulb",
			RunE: func(cmd *cobra.Command, args []string) error {
				colorErr := fmt.Errorf("please provide a color name or hex color")
				if len(args) < 1 {
					return colorErr
				}
				color := control.ParseColorString(args[0])
				if color == nil {
					return colorErr
				}
				controller.SetColor(*color)
				return nil
			},
		},
		{
			Use:   "set-power",
			Short: "turn the lightbulb 'on' or 'off'",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("please specify 'on' or 'off'")
				}
				if args[0] == "on" {
					controller.SetPower(true)
				} else if args[0] == "off" {
					controller.SetPower(false)
				} else {
					return fmt.Errorf("please specify 'on' or 'off'")
				}
				return nil
			},
		}, cobra.Command{
			Use:   "get-state",
			Short: "get the state of the lightbulb",
			RunE: func(cmd *cobra.Command, args []string) error {
				state, err := controller.GetState()
				if err != nil {
					return err
				}
				fmt.Printf("IsOn: %t\n", state.IsOn)
				fmt.Printf("DeviceType: %x\n", state.DeviceType)
				fmt.Printf("LedVersionNum: %d\n", state.LedVersionNum)
				fmt.Printf("Mode: %s\n", control.ModeName(state.Mode))
				fmt.Printf("Slowness: %d\n", state.Slowness)
				fmt.Printf("Color: %s\n", control.ColorToStr(state.Color))
				return nil
			},
		}, cobra.Command{
			Use:   "get-time",
			Short: "get the current time from the point of view of the lightbulb",
			RunE: func(cmd *cobra.Command, args []string) error {
				time, err := controller.GetTime()
				if err != nil {
					return err
				}
				fmt.Printf("Time: %s\n", time.Time)
				return nil
			},
		}, cobra.Command{
			Use:   "get-timers",
			Short: "get the list of timers set on the lightbulb",
			RunE: func(cmd *cobra.Command, args []string) error {
				timers, err := controller.GetTimers()
				if err != nil {
					return err
				}
				for i, timer := range timers.Timers {
					fmt.Printf("Timer %d\n", i)
					fmt.Printf("Enabled: %t\n", timer.Enabled)
					fmt.Printf("PowerOn: %t\n", timer.PowerOn)
					fmt.Printf("Mode: %s\n", control.ModeName(timer.Mode))
					fmt.Printf("Time: %s\n", timer.Time)
					fmt.Print("Weekdays:")
					for _, d := range timer.Weekdays {
						if d {
							fmt.Print("+")
						} else {
							fmt.Print("-")
						}
					}
					fmt.Print("\n")
					fmt.Printf("Data: %x\n", timer.Data)
				}
				return nil
			},
		}, cobra.Command{
			Use:   "atmolight",
			Short: "use atmolight feature",
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("Please provide the path to a pty")
				}
				return controller.AtmolightDaemon(args[0])
			},
		},
	}

	zenggeCmd.AddCommand(localCmd)
	zenggeCmd.AddCommand(manageCmd)
	zenggeCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(commandCmd)
	addAll(manageCmd, manageCmds)
	addAll(remoteCmd, remoteCmds)
	addAll(localCmd, execCmds)
	// we must copy the structs if using them for more than one path
	addAll(commandCmd, execCmds[:])

	zenggeCmd.Execute()
}
