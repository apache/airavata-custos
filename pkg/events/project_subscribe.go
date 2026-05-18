package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ProjectHandler handles project lifecycle events with a typed payload.
type ProjectHandler func(project models.Project)

// SubscribeProjectCreated registers a typed handler invoked whenever a
// project::create event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectCreated(handler ProjectHandler) {
	b.subscribeProject(ProjectCreateEvent, handler)
}

// SubscribeProjectUpdated registers a typed handler invoked whenever a
// project::update event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectUpdated(handler ProjectHandler) {
	b.subscribeProject(ProjectUpdateEvent, handler)
}

// SubscribeProjectDeleted registers a typed handler invoked whenever a
// project::delete event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectDeleted(handler ProjectHandler) {
	b.subscribeProject(ProjectDeleteEvent, handler)
}

func (b *Bus) subscribeProject(topic EventType, handler ProjectHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch p := value.(type) {
		case models.Project:
			handler(p)
		case *models.Project:
			if p != nil {
				handler(*p)
			}
		default:
			slog.Warn("project event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
