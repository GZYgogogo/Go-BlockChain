package core

import "fmt"

type Status struct {
	data map[string][]byte
}

func NewStatus() *Status {
	return &Status{
		data: make(map[string][]byte),
	}
}

func (s *Status) Put(key, value []byte) error {
	s.data[string(key)] = value
	return nil
}

func (s *Status) Delete(key string) error {
	delete(s.data, key)
	return nil
}

func (s *Status) Get(key string) ([]byte, error) {
	value, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return value, nil
}
