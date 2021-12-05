package core

import (
	"encoding/json"
	"github.com/hashicorp/raft"
)

func (s *Store) Get(k string, level ConsistencyLevel) (string, error) {
	if level != Stale {
		if s.raft.State() != raft.Leader {
			return "", ErrNotLeader
		}
	}

	if level == Consistent {
		if err := s.consistentRead(); err != nil {
			return "", err
		}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.m[k], nil
}

func (s *Store) Set(k, v string) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := &command{
		Op:    "set",
		Key:   k,
		Value: v,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (s *Store) Delete(key string) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	c := &command{
		Op:  "delete",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}



