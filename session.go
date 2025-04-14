package simplehttp

// Session defines the interface for session management
type Session interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	Delete(key string) error
	Clear() error
	ID() string
	Save() error
}

// MemorySession provides a simple in-memory session implementation
type MemorySession struct {
	id   string
	data map[string]interface{}
}

func NewMemorySession(id string) Session {
	return &MemorySession{
		id:   id,
		data: make(map[string]interface{}),
	}
}

func (s *MemorySession) Get(key string) interface{} {
	return s.data[key]
}

func (s *MemorySession) Set(key string, value interface{}) error {
	s.data[key] = value
	return nil
}

func (s *MemorySession) Delete(key string) error {
	delete(s.data, key)
	return nil
}

func (s *MemorySession) Clear() error {
	s.data = make(map[string]interface{})
	return nil
}

func (s *MemorySession) ID() string {
	return s.id
}

func (s *MemorySession) Save() error {
	// In memory implementation doesn't need to save
	return nil
}
