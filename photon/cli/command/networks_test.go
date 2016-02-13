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

type MockNetworksPage struct {
	Items            []photon.Network `json:"items"`
	NextPageLink     string           `json:"nextPageLink"`
	PreviousPageLink string           `json:"previousPageLink"`
}

func TestCreateDeleteNetwork(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_NETWORK",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/networks",
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
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	set.String("name", "network_name", "network name")
	set.String("portgroups", "portgroup, portgroup2", "portgroups")

	cxt := cli.NewContext(nil, set, globalCtx)

	err = createNetwork(cxt)
	if err != nil {
		t.Error("Not expecting create network to fail", err)
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "network-ID"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/networks/"+queuedTask.Entity.ID,
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
	cxt = cli.NewContext(nil, set, globalCtx)
	err = deleteNetwork(cxt)
	if err != nil {
		t.Error("Not expecting delete network to fail")
	}
}

func TestListNetworks(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	expectedList := MockNetworksPage{
		Items: []photon.Network{
			photon.Network{
				ID:         "network_id",
				Name:       "network_name",
				PortGroups: []string{"port", "group"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/networks",
		mocks.CreateResponder(200, string(response[:])))

	expectedList = MockNetworksPage{
		Items:            []photon.Network{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)

	cxt := cli.NewContext(nil, set, nil)
	err = listNetworks(cxt)
	if err != nil {
		t.Error("Error listing networks: " + err.Error())
	}
}

func TestShowNetworks(t *testing.T) {
	expectedStruct := photon.Network{
		ID:         "network_id",
		Name:       "network_name",
		PortGroups: []string{"port", "group"},
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/networks/"+expectedStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedStruct.ID})
	cxt := cli.NewContext(nil, set, nil)
	err = showNetwork(cxt)
	if err != nil {
		t.Error("Error showing networks: " + err.Error())
	}
}