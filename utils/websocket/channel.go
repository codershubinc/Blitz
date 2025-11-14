package websocket

import (
	"Blitz/models"
	"log"
	"sync"
)

var (
	sharedChannel chan models.ServerResponse
	once          sync.Once
	mu            sync.RWMutex
)

func CreateChannel() chan models.ServerResponse {
	once.Do(func() {
		sharedChannel = make(chan models.ServerResponse, 100)
		log.Println("Channel created successfully")
	})
	return sharedChannel
}

func GetChannel() chan models.ServerResponse {
	mu.RLock()
	defer mu.RUnlock()
	if sharedChannel == nil {
		sharedChannel = make(chan models.ServerResponse)
		log.Println("Channel created inside GetChannel")
	} else {
		log.Println("Channel already exists, returning existing channel")
	}
	return sharedChannel
}
func CloseChannel() {
	mu.Lock()
	defer mu.Unlock()
	if sharedChannel != nil {
		close(sharedChannel)
		sharedChannel = nil
		log.Println("Channel closed and set to nil")
	} else {
		log.Println("Channel is already nil, nothing to close")
	}
}

func WriteChannelMessage(msg models.ServerResponse) {
	mu.RLock()
	ch := sharedChannel
	defer mu.RUnlock()


	if ch == nil {
		log.Println("Channel is nil, cannot send message")
		return
	}

	select {
	case ch <- msg:
		// log.Println("Message sent to channel:", msg)
	default:
		log.Println("Channel is full, message not sent:", msg)
	}

}
