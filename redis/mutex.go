package redis

import (
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/rickone/athena/errcode"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

const (
	mutexDefaultLockTimeout = 6 * time.Second
	mutexDefaultLeaseTime   = 20 * time.Second
	backoffTime             = 2 * time.Millisecond
)

type Mutex struct {
	conn   redigo.Conn
	key    string
	value  string
	option MutexOption
}

type MutexOption struct {
	LockTimeout time.Duration
	LeaseTime   time.Duration
}

func NewMutex(key string, opts ...MutexOption) *Mutex {
	m := &Mutex{
		conn:  DB("mutex").Pool.Get(),
		key:   key,
		value: uuid.New().String(),
	}
	if len(opts) > 0 {
		m.option = opts[0]
	} else {
		m.option = MutexOption{
			LockTimeout: mutexDefaultLockTimeout,
			LeaseTime:   mutexDefaultLeaseTime,
		}
	}
	return m
}

func (m *Mutex) Close() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
}

func (m *Mutex) TryLock() error {
	_, err := redigo.String(m.conn.Do("SET", m.key, m.value, "EX", m.option.LeaseTime.Seconds()+2, "NX"))
	if err == redigo.ErrNil {
		return status.Errorf(errcode.ErrMutexLock, "lock '%s' failed", m.key)
	}
	return err
}

func (m *Mutex) Lock() error {
	total := time.Duration(0)
	s := backoffTime
	for {
		err := m.TryLock()
		if err == nil {
			return nil
		}
		if err != redigo.ErrNil {
			return err
		}

		time.Sleep(s)
		total += s
		if total > m.option.LockTimeout {
			return status.Errorf(errcode.ErrMutexLock, "lock '%s' failed", m.key)
		}

		s *= 2
	}
}

func (m *Mutex) Unlock() error {
	value, err := redigo.String(m.conn.Do("GET", m.key))
	if err != nil {
		return err
	}
	if value != m.value {
		return status.Errorf(errcode.ErrMutexUnlock, "unlock '%s' failed", m.key)
	}

	ttl, err := redigo.Int(m.conn.Do("TTL", m.key))
	if err != nil {
		return err
	}

	if ttl > 1 {
		_, err = m.conn.Do("DEL", m.key)
		return err
	}

	return nil
}

func MutexWrap(key string, f func() (interface{}, error)) (interface{}, error) {
	m := NewMutex(key)
	defer m.Close()

	if err := m.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		err := m.Unlock()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"key": key,
				"err": err.Error(),
			}).Error("Mutex unlock failed")
		}
	}()

	return f()
}
