package manager

import "github.com/campadrenalin/go-deje/model"

type EventManager struct {
	ObjectManager
}

func NewEventManager() EventManager {
	om := NewObjectManager()
	return EventManager{om}
}

func (em *EventManager) Register(event model.Event) {
	em.register(event)
}

func (em *EventManager) Unregister(event model.Event) {
	em.unregister(event)
}
