package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/pkg/common"
	"github.com/keptn-contrib/dynatrace-service/pkg/credentials"
	keptnutils "github.com/keptn/go-utils/pkg/api/utils"

	keptn "github.com/keptn/go-utils/pkg/lib"
)

const DefaultOperatorVersion = "v0.8.0"
const sliResourceURI = "dynatrace/sli.yaml"
const Throughput = "throughput"
const ErrorRate = "error_rate"
const ResponseTimeP50 = "response_time_p50"
const ResponseTimeP90 = "response_time_p90"
const ResponseTimeP95 = "response_time_p95"

type criteriaObject struct {
	Operator        string
	Value           float64
	CheckPercentage bool
	IsComparison    bool
	CheckIncrease   bool
}

type DTAPIListResponse struct {
	Values []Values `json:"values"`
}
type Values struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DynatraceHelper struct {
	DynatraceCreds *credentials.DTCredentials
	Logger         keptn.LoggerInterface
	OperatorTag    string
	KeptnHandler   *keptn.Keptn
	KeptnBridge    string
}

// NewDynatraceHelper creates a new DynatraceHelper
func NewDynatraceHelper(keptnHandler *keptn.Keptn, dynatraceCreds *credentials.DTCredentials, logger keptn.LoggerInterface) *DynatraceHelper {
	return &DynatraceHelper{
		DynatraceCreds: dynatraceCreds,
		KeptnHandler:   keptnHandler,
		Logger:         logger,
	}
}

// ConfigureMonitoring configures Dynatrace for a Keptn project
func (dt *DynatraceHelper) ConfigureMonitoring(project string, shipyard *keptn.Shipyard) error {

	dt.EnsureDTTaggingRulesAreSetUp()

	dt.EnsureProblemNotificationsAreSetUp()

	if project != "" && shipyard != nil {
		dt.CreateManagementZones(project, *shipyard)

		configHandler := keptnutils.NewServiceHandler("configuration-service:8080")
		dt.CreateDashboard(project, *shipyard)

		// try to create metric events - if one fails, don't fail the whole setup
		for _, stage := range shipyard.Stages {
			if stage.RemediationStrategy == "automated" {
				services, err := configHandler.GetAllServices(project, stage.Name)
				if err != nil {
					return fmt.Errorf("failed to retrieve services of project %s: %v", project, err.Error())
				}
				for _, service := range services {
					dt.CreateMetricEvents(project, stage.Name, service.ServiceName)
				}
			}
		}
	}
	return nil
}

/**
 * if dtCredsSecretName is passed and it is not dynatrace (=default) then we try to pull the secret based on that name and is it for this API Call
 */
func (dt *DynatraceHelper) sendDynatraceAPIRequest(apiPath string, method string, body []byte) (string, error) {

	if common.RunLocal || common.RunLocalTest {
		dt.Logger.Info("Dynatrace.sendDynatraceAPIRequest(RUNLOCAL) - not sending event to " +
			dt.DynatraceCreds.Tenant + "). Here is the payload: " + string(body))
		return "", nil
	}

	var url string
	if !strings.HasPrefix(dt.DynatraceCreds.Tenant, "http://") && !strings.HasPrefix(dt.DynatraceCreds.Tenant, "https://") {
		url = "https://" + dt.DynatraceCreds.Tenant + apiPath
	} else {
		url = dt.DynatraceCreds.Tenant + apiPath
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Token "+dt.DynatraceCreds.ApiToken)
	req.Header.Set("User-Agent", "keptn-contrib/dynatrace-service:"+os.Getenv("version"))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !IsHttpSSLVerificationEnabled()},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send Dynatrace API request: %v", err)
	}

	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(responseBody), fmt.Errorf("api request failed with status %s and response %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}
