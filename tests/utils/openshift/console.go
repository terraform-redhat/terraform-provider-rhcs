package openshift

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	client "github.com/openshift-online/ocm-sdk-go"
	CMS "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// Console is openshift webconsole struct
type Console struct {
	ClusterName string
	ClusterID   string
	KubePath    string
	Private     bool
}

// NewConsole need clusterID and connection passed. Will return the console with login=true
func NewConsole(clusterID string, connection *client.Connection) (console *Console, err error) {
	console = &Console{ClusterID: clusterID}

	// detailResp, err := CMS.RetrieveClusterDetail(connection, clusterID)
	if err != nil {
		return
	}
	if console.IsConfiged() {
		return console, nil
	}
	resp, err := CMS.RetrieveClusterCredentials(connection, clusterID)
	if err != nil {
		return
	}
	if resp.Status() != http.StatusOK {
		err = errors.New(fmt.Sprintf("/api/clusters_mgmt/v1/clusters/%s/credentials request failed, ERROR:\n %s", clusterID, err))
	}
	kubeConfig := resp.Body().Kubeconfig()
	if kubeConfig == "" {
		return
	}
	console = &Console{ClusterID: clusterID}
	console, err = console.Config(kubeConfig)
	return
}

// IsConfiged will return whether the config file of the console is existed
func (c *Console) IsConfiged() bool {
	configFile := filepath.Join(CON.GetRHCSOutputDir(), h.Join(c.ClusterID, CON.ConfigSuffix))
	_, err := os.Stat(configFile)
	if err != nil {
		return false
	}
	c.KubePath = configFile
	return true
}

// Config will return the console which had been configured
func (c *Console) Config(kubeConfig string) (*Console, error) {
	configFile := filepath.Join(CON.GetRHCSOutputDir(), h.Join(c.ClusterID, CON.ConfigSuffix))
	wf, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0766)
	if err == nil {
		_, err = wf.Write([]byte(kubeConfig))
		wf.Close()
	}
	c.KubePath = configFile
	return c, err
}
