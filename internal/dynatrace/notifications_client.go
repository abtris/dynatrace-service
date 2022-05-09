package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keptn-contrib/dynatrace-service/internal/credentials"
)

const keptnProblemNotificationName = "Keptn Problem Notification"

const problemNotificationPayload string = `{ 
      "type": "WEBHOOK", 
      "name": "$KEPTN_PROBLEM_NOTIFICATION_NAME", 
      "alertingProfile": "$ALERTING_PROFILE_ID", 
      "active": true, 
      "url": "$KEPTN_DNS/v1/event", 
      "acceptAnyCertificate": true, 
      "headers": [ 
        { "name": "x-token", "value": "$KEPTN_TOKEN" },
        { "name": "Content-Type", "value": "application/cloudevents+json" }
      ],
      "payload": "{\n    \"specversion\":\"1.0\",\n    \"type\":\"sh.keptn.events.problem\",\n    \"shkeptncontext\":\"{PID}\",\n    \"source\":\"dynatrace\",\n    \"id\":\"{PID}\",\n    \"time\":\"\",\n    \"contenttype\":\"application/json\",\n    \"data\": {\n        \"State\":\"{State}\",\n        \"ProblemID\":\"{ProblemID}\",\n        \"PID\":\"{PID}\",\n        \"ProblemTitle\":\"{ProblemTitle}\",\n        \"ProblemURL\":\"{ProblemURL}\",\n        \"ProblemDetails\":{ProblemDetailsJSON},\n        \"Tags\":\"{Tags}\",\n        \"ImpactedEntities\":{ImpactedEntities},\n        \"ImpactedEntity\":\"{ImpactedEntity}\",\n        \"KeptnProject\":\"$KEPTN_PROJECT\"\n    }\n}\n" 

      }`

type NotificationsError struct {
	errors []error
}

func (ne *NotificationsError) HasErrors() bool {
	return len(ne.errors) > 0
}

func (ne *NotificationsError) Error() string {
	sb := strings.Builder{}
	for i, err := range ne.errors {
		sb.WriteString(err.Error())
		if i < len(ne.errors)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

const notificationsPath = "/api/config/v1/notifications"

type NotificationsClient struct {
	client ClientInterface
}

func NewNotificationsClient(client ClientInterface) *NotificationsClient {
	return &NotificationsClient{
		client: client,
	}
}

func (nc *NotificationsClient) getAll(ctx context.Context) (*listResponse, error) {
	response, err := nc.client.Get(ctx, notificationsPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve notifications: %v", err)
	}

	existingNotifications := &listResponse{}
	err = json.Unmarshal(response, existingNotifications)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal notifications: %v", err)
	}

	return existingNotifications, nil
}

// DeleteExistingKeptnProblemNotifications deletes all existing Keptn problem notifications.
func (nc *NotificationsClient) DeleteExistingKeptnProblemNotifications(ctx context.Context) error {
	existingNotifications, err := nc.getAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve notifications: %v", err)
	}

	notificationError := &NotificationsError{}
	for _, notification := range existingNotifications.Values {
		if notification.Name == keptnProblemNotificationName {
			err := nc.deleteBy(ctx, notification.ID)
			if err != nil {
				// Error occurred but continue
				notificationError.errors = append(
					notificationError.errors,
					fmt.Errorf("failed to delete notification with ID: %s", notification.ID))
			}
		}
	}

	if notificationError.HasErrors() {
		return notificationError
	}

	return nil
}

// Create creates a new default notification for the given KeptnAPICredentials and the alertingProfileID.
func (nc *NotificationsClient) Create(ctx context.Context, credentials *credentials.KeptnCredentials, alertingProfileID string, project string) error {
	notification := problemNotificationPayload
	notification = strings.ReplaceAll(notification, "$KEPTN_DNS", credentials.GetAPIURL())
	notification = strings.ReplaceAll(notification, "$KEPTN_TOKEN", credentials.GetAPIToken())
	notification = strings.ReplaceAll(notification, "$ALERTING_PROFILE_ID", alertingProfileID)
	notification = strings.ReplaceAll(notification, "$KEPTN_PROBLEM_NOTIFICATION_NAME", keptnProblemNotificationName)
	notification = strings.ReplaceAll(notification, "$KEPTN_PROJECT", project)

	_, err := nc.client.Post(ctx, notificationsPath, []byte(notification))
	if err != nil {
		return err
	}

	return nil
}

func (nc *NotificationsClient) deleteBy(ctx context.Context, id string) error {
	_, err := nc.client.Delete(ctx, notificationsPath+"/"+id)
	if err != nil {
		return nil
	}

	return nil
}
