package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.73

import (
	"context"
	"fmt"
	database_models "go-cron-webhooks/database-models"
	"go-cron-webhooks/graph/model"
	"time"
)

// CreateCronWebhook is the resolver for the createCronWebhook field.
func (r *mutationResolver) CreateCronWebhook(ctx context.Context, url string, cronExpression string) (*model.CronWebhook, error) {
	newWebhook := &database_models.CronWebhookModel{
		Cron:    cronExpression,
		URL:     url,
		Enabled: true, // default to enabled
	}

	if err := r.DB.Create(newWebhook).Error; err != nil {
		return nil, fmt.Errorf("failed to create cron webhook: %w", err)
	}
	// Create the model.CronWebhook from the database model

	// Add to the cron runner
	cronId, err := r.C.AddFunc(newWebhook.Cron, func() {
		database_models.RunWebhook(r.DB, newWebhook)
	})

	nextRun := r.C.Entry(cronId).Next.Format(time.RFC3339)

	if err != nil {
		// delete the webhook from the database if cron job creation fails
		if dbErr := r.DB.Delete(newWebhook).Error; dbErr != nil {
			return nil, fmt.Errorf("failed to delete cron webhook after cron job creation failure: %w", dbErr)
		}

		return nil, fmt.Errorf("failed to add cron job for webhook %d: %w", newWebhook.ID, err)
	}

	createdWebhook := &model.CronWebhook{
		ID:             fmt.Sprintf("%d", newWebhook.ID),
		URL:            newWebhook.URL,
		CronExpression: newWebhook.Cron,
		Enabled:        newWebhook.Enabled,
		NextRun:        &nextRun,
	}

	r.WebhookMap[newWebhook.ID] = cronId

	// return the created webhook
	return createdWebhook, nil
}

// TriggerWebhook is the resolver for the triggerWebhook field.
func (r *mutationResolver) TriggerWebhook(ctx context.Context, id string) (bool, error) {
	// Find the existing webhook by ID
	var existingWebhook database_models.CronWebhookModel
	if err := r.DB.First(&existingWebhook, id).Error; err != nil {
		return false, fmt.Errorf("failed to find cron webhook with ID %s: %w", id, err)
	}

	// Send HTTP request to the webhook URL
	err := database_models.RunWebhook(r.DB, &existingWebhook)
	if err != nil {
		return false, fmt.Errorf("failed to trigger webhook %s: %w", existingWebhook.URL, err)
	}
	// Log successful request
	fmt.Printf("successfully triggered webhook %s\n", existingWebhook.URL)

	return true, nil
}

// UpdateCronWebhook is the resolver for the updateCronWebhook field.
func (r *mutationResolver) UpdateCronWebhook(ctx context.Context, id string, url *string, cronExpression *string, enabled *bool) (*model.CronWebhook, error) {
	// Find the existing webhook by ID
	var existingWebhook database_models.CronWebhookModel
	if err := r.DB.First(&existingWebhook, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find cron webhook with ID %s: %w", id, err)
	}

	// Update fields if provided
	if url != nil {
		existingWebhook.URL = *url
	}
	if cronExpression != nil {
		existingWebhook.Cron = *cronExpression
	}
	if enabled != nil {
		existingWebhook.Enabled = *enabled
	}

	// Save the updated webhook back to the database
	if err := r.DB.Save(&existingWebhook).Error; err != nil {
		return nil, fmt.Errorf("failed to update cron webhook: %w", err)
	}

	// Remove the existing entry from the cron runner if it exists
	if entryID, exists := r.WebhookMap[existingWebhook.ID]; exists {
		r.C.Remove(entryID)
		delete(r.WebhookMap, existingWebhook.ID)
	}

	updatedWebhook := &model.CronWebhook{
		ID:             fmt.Sprintf("%d", existingWebhook.ID),
		URL:            existingWebhook.URL,
		CronExpression: existingWebhook.Cron,
		Enabled:        existingWebhook.Enabled,
		NextRun:        nil, // NextRun will be set if the webhook is enabled
	}
	// re-create webhook if enabled
	if existingWebhook.Enabled {
		newEntryId, err := r.C.AddFunc(existingWebhook.Cron, func() {
			database_models.RunWebhook(r.DB, &existingWebhook)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add cron job for webhook %d: %w", existingWebhook.ID, err)
		}
		// Store the entry ID in the WebhookMap
		r.WebhookMap[existingWebhook.ID] = newEntryId
		nextRunIso := r.C.Entry(newEntryId).Next.Format(time.RFC3339)
		updatedWebhook.NextRun = &nextRunIso
	}

	// Create the model.CronWebhook from the updated database model

	return updatedWebhook, nil
}

// DeleteCronWebhook is the resolver for the deleteCronWebhook field.
func (r *mutationResolver) DeleteCronWebhook(ctx context.Context, id string) (bool, error) {
	// Find the existing webhook by ID
	var existingWebhook database_models.CronWebhookModel
	if err := r.DB.First(&existingWebhook, id).Error; err != nil {
		return false, fmt.Errorf("failed to find cron webhook with ID %s: %w", id, err)
	}

	// if enabled, remove from cron runner
	if existingWebhook.Enabled {
		if entryID, exists := r.WebhookMap[existingWebhook.ID]; exists {
			r.C.Remove(entryID)
			delete(r.WebhookMap, existingWebhook.ID)
		}
	}

	// Delete the webhook from the database
	if err := r.DB.Delete(&existingWebhook).Error; err != nil {
		return false, fmt.Errorf("failed to delete cron webhook: %w", err)
	}

	// Return true to indicate successful deletion
	return true, nil
}

// CronWebhooks is the resolver for the cronWebhooks field.
func (r *queryResolver) CronWebhooks(ctx context.Context) ([]*model.CronWebhook, error) {
	// Fetch all cron webhooks from the database
	var webhooks []database_models.CronWebhookModel
	if err := r.DB.Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch cron webhooks: %w", err)
	}

	// Convert database models to GraphQL models
	var result []*model.CronWebhook
	for _, webhook := range webhooks {

		// get entry ID from the cron runner if it exists
		webhookModel := &model.CronWebhook{
			ID:             fmt.Sprintf("%d", webhook.ID),
			URL:            webhook.URL,
			CronExpression: webhook.Cron,
			Enabled:        webhook.Enabled,
			NextRun:        nil, // NextRun will be set if the webhook is enabled
		}

		if entryID, exists := r.WebhookMap[webhook.ID]; exists {
			nextRunIso := r.C.Entry(entryID).Next.Format(time.RFC3339)
			webhookModel.NextRun = &nextRunIso
		}

		result = append(result, webhookModel)
	}

	return result, nil
}

// CronWebhook is the resolver for the cronWebhook field.
func (r *queryResolver) CronWebhook(ctx context.Context, id string) (*model.CronWebhook, error) {
	// Find the cron webhook by ID
	var webhook database_models.CronWebhookModel
	if err := r.DB.First(&webhook, id).Error; err != nil {
		return nil, fmt.Errorf("failed to find cron webhook with ID %s: %w", id, err)
	}

	webhookModel := &model.CronWebhook{
		ID:             fmt.Sprintf("%d", webhook.ID),
		URL:            webhook.URL,
		CronExpression: webhook.Cron,
		Enabled:        webhook.Enabled,
		NextRun:        nil, // NextRun will be set if the webhook is enabled
	}

	if entryID, exists := r.WebhookMap[webhook.ID]; exists {
		nextRunIso := r.C.Entry(entryID).Next.Format(time.RFC3339)
		webhookModel.NextRun = &nextRunIso
	}

	// Convert the database model to GraphQL model
	return webhookModel, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
