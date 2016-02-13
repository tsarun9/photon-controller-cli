package command

import (
	"encoding/json"
	"flag"
	"net/http"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	"github.com/vmware/photon-controller-cli/photon/cli/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

type MockHostsPage struct {
	Items            []photon.Host `json:"items"`
	NextPageLink     string        `json:"nextPageLink"`
	PreviousPageLink string        `json:"previousPageLink"`
}

func TestListDeployment(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	err := set.Parse([]string{""})
	cxt := cli.NewContext(nil, set, nil)
	err = listDeployments(cxt)
	// No responder from mock server for list tenant set yet
	if err == nil {
		t.Error("Expecting an error listing deployments")
	}
}

func TestCreateDeleteDeployment(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_DEPLOYMENT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_DEPLOYMENT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	globalCtx := cli.NewContext(nil, globalSet, nil)
	set := flag.NewFlagSet("test", 0)
	set.String("image_datastores", "testname", "name")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createDeployment(cxt)
	if err != nil {
		t.Error("Not expecting create Deployment to fail")
	}

	expectedStruct := photon.Deployments{
		Items: []photon.Deployment{
			photon.Deployment{
				ImageDatastores: []string{"testname"},
				ID:              "1",
			},
			photon.Deployment{
				ImageDatastores: []string{"secondname"},
				ID:              "2",
			},
		},
	}

	response, err = json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{})
	cxt = cli.NewContext(nil, set, nil)
	err = listDeployments(cxt)
	if err != nil {
		t.Error("Not expecting list deployment to fail")
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_HOST",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_HOST",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/deployments/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)
	err = deleteDeployment(cxt)
	if err != nil {
		t.Error("Not expecting delete Deployment to fail")
	}
}

func TestGetDeployment(t *testing.T) {
	auth := &photon.AuthInfo{
		Enabled: false,
	}
	getStruct := photon.Deployment{
		ImageDatastores: []string{"testname"},
		ID:              "1",
		Auth:            auth,
		State:           "COMPLETED",
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/"+getStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{getStruct.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showDeployment(cxt)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestListDeploymentHosts(t *testing.T) {
	hostList := MockHostsPage{
		Items: []photon.Host{
			photon.Host{
				Username: "u",
				Password: "p",
				Address:  "testIP",
				Tags:     []string{"CLOUD"},
				ID:       "host-test-id",
				State:    "COMPLETED",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(hostList)
	if err != nil {
		t.Error("Not expecting error serializing host list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/1/hosts",
		mocks.CreateResponder(200, string(response[:])))

	hostList = MockHostsPage{
		Items:            []photon.Host{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	response, err = json.Marshal(hostList)
	if err != nil {
		t.Error("Not expecting error serializing expected taskLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = listDeploymentHosts(cxt)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestListDeploymentVms(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	vmList := MockVMsPage{
		Items: []photon.VM{
			photon.VM{
				Name:          "fake_vm_name",
				ID:            "fake_vm_ID",
				Flavor:        "fake_vm_flavor_name",
				State:         "STOPPED",
				SourceImageID: "fake_image_ID",
				Host:          "fake_host_ip",
				Datastore:     "fake_datastore_ID",
				AttachedDisks: []photon.AttachedDisk{
					photon.AttachedDisk{
						Name:       "d1",
						Kind:       "ephemeral-disk",
						Flavor:     "fake_ephemeral_flavor_ID",
						CapacityGB: 0,
						BootDisk:   true,
					},
				},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializing vm list")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/1/vms",
		mocks.CreateResponder(200, string(response[:])))

	vmList = MockVMsPage{
		Items:            []photon.VM{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializing vm list")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = listDeploymentVms(cxt)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestInitializeMigrateDeployment(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "INITIALIZE_MIGRATE_DEPLOYMENT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "INITIALIZE_MIGRATE_DEPLOYMENT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/1/initialize_migration",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000", "1"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = initializeMigrateDeployment(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting initialize Deployment to fail")
	}
}

func TestFinalizeeMigrateDeployment(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "FINALIZE_MIGRATE_DEPLOYMENT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "FINALIZE_MIGRATE_DEPLOYMENT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/1/finalize_migration",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000", "1"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = finalizeMigrateDeployment(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting initialize Deployment to fail")
	}
}

func TestUpdateImageDatastores(t *testing.T) {
	deploymentId := "deployment1"
	completedTask := photon.Task{
		ID:        "task1",
		Operation: "UPDATE_IMAGE_DATASTORES",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error when serializing tasks")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/"+deploymentId+"/set_image_datastores",
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{deploymentId, "ds1,ds2"})
	if err != nil {
		t.Error(err)
	}
	ctx := cli.NewContext(nil, set, nil)

	err = updateImageDatastores(ctx)
	if err != nil {
		t.Error(err)
	}
}