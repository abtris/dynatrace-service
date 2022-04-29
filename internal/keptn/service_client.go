package keptn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/keptn-contrib/dynatrace-service/internal/rest"
	apimodels "github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
)

// ServiceClientInterface provides access to Keptn services.
type ServiceClientInterface interface {
	// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
	GetServiceNames(ctx context.Context, project string, stage string) ([]string, error)

	// CreateServiceInProject creates a service in all stages of the specified project or returns an error.
	CreateServiceInProject(ctx context.Context, project string, service string) error
}

// ServiceClient is an implementation of ServiceClientInterface using api.ServicesV1Interface and APIClientInterface.
type ServiceClient struct {
	servicesClient api.ServicesV1Interface
	apiClient      APIClientInterface
}

// NewServiceClient creates a new ServiceClient using the specified clients.
func NewServiceClient(servicesClient api.ServicesV1Interface, httpClient *http.Client) *ServiceClient {
	return &ServiceClient{
		servicesClient: servicesClient,
		apiClient: NewAPIClient(
			rest.NewDefaultClient(
				httpClient,
				getShipyardControllerURL())),
	}
}

// GetServiceNames gets the names of the services in the specified project and stage or returns an error.
func (c *ServiceClient) GetServiceNames(ctx context.Context, project string, stage string) ([]string, error) {
	services, err := c.servicesClient.GetAllServicesWithContext(ctx, project, stage)
	if err != nil {
		return nil, fmt.Errorf("could not fetch services of Keptn project %s at stage %s: %s", project, stage, err.Error())
	}

	if services == nil {
		return nil, nil
	}

	serviceNames := make([]string, len(services))
	for i, service := range services {
		serviceNames[i] = service.ServiceName
	}

	return serviceNames, nil
}

// CreateServiceInProject creates a service in all stages of the specified project or returns an error.
func (c *ServiceClient) CreateServiceInProject(ctx context.Context, project string, service string) error {
	serviceModel := &apimodels.CreateService{
		ServiceName: &service,
	}
	reqBody, err := json.Marshal(serviceModel)
	if err != nil {
		return fmt.Errorf("could not marshal service payload: %s", err.Error())
	}

	_, err = c.apiClient.Post(ctx, getServicePathFor(project), reqBody)
	return err
}

func getServicePathFor(project string) string {
	return "/v1/project/" + project + "/service"
}
