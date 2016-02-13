package command

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/cli/client"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for availability-zone
// Subcommands: create; Usage: availability-zone create <name>
//              delete; Usage: availability-zone delete <id>
//              list;   Usage: availability-zone list
//              show;   Usage: availability-zone show <id>
//              tasks;  Usage: availability-zone tasks <id> [<options>]
func GetAvailabilityZonesCommand() cli.Command {
	command := cli.Command{
		Name:  "availability-zone",
		Usage: "options for availability-zone",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create new availability-zone",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Availability-zone name",
					},
				},
				Action: func(c *cli.Context) {
					err := createAvailabilityZone(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete availability-zone",
				Action: func(c *cli.Context) {
					err := deleteAvailabilityZone(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List availability-zones",
				Action: func(c *cli.Context) {
					err := listAvailabilityZones(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show specified availability-zone",
				Action: func(c *cli.Context) {
					err := showAvailabilityZone(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "tasks",
				Usage: "Show availability-zone tasks",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task sate",
					},
				},
				Action: func(c *cli.Context) {
					err := getAvailabilityZoneTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create availability-zone task to client based on the cli.Context
// Returns an error if one occurred
func createAvailabilityZone(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Unknown argument: %v", c.Args()[1:])
	}
	name := c.Args().First()

	if !c.GlobalIsSet("non-interactive") {
		var err error
		name, err = askForInput("AvailabilityZone name: ", name)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide availability zone name")
	}

	azSpec := &photon.AvailabilityZoneCreateSpec{
		Name: name,
	}

	var err error
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	createTask, err := client.Esxclient.AvailabilityZones.Create(azSpec)

	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		task, err := client.Esxclient.Tasks.Wait(createTask.ID)
		if err != nil {
			return nil
		}
		fmt.Printf("%s\t%s\n", name, task.Entity.ID)
	} else {
		task, err := pollTask(createTask.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Created availability zone '%s' ID: %s \n", name, task.Entity.ID)
	}
	return nil
}

// Retrieves availability zone against specified id.
func showAvailabilityZone(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "availability-zone show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	zone, err := client.Esxclient.AvailabilityZones.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\n", zone.ID, zone.Name, zone.State)
	} else {
		fmt.Printf("AvailabilityZone ID: %s\n", zone.ID)
		fmt.Printf("  Name:        %s\n", zone.Name)
		fmt.Printf("  State:       %s\n", zone.State)
	}

	return nil
}

// Retrieves a list of availability zones, returns an error if one occurred
func listAvailabilityZones(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "availability zone list")
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	zones, err := client.Esxclient.AvailabilityZones.GetAll()
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Println(len(zones.Items))
		for _, zone := range zones.Items {
			fmt.Printf("%s\t%s\n", zone.ID, zone.Name)
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\n")
		for _, zone := range zones.Items {
			fmt.Fprintf(w, "%s\t%s\n", zone.ID, zone.Name)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(zones.Items))
	}

	return nil
}

// Sends a delete availability zone task to client based on the cli.Context
// Returns an error if one occurred
func deleteAvailabilityZone(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "availability zone delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.AvailabilityZones.Delete(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(deleteTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Retrieves tasks from specified availability zone
func getAvailabilityZoneTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "availability zone task <id> [<options>]")
	if err != nil {
		return err
	}
	id := c.Args().First()

	state := c.String("state")
	options := &photon.TaskGetOptions{
		State: state,
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	taskList, err := client.Esxclient.AvailabilityZones.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	return nil
}