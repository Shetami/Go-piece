package main

import (
	"sync"
	"time"
)

type session struct {
	user    string
	expires time.Time
}

// Store — сессии пользователей с временем жизни. Get зовут на каждый
// запрос, Set — при логине, поэтому чтений намного больше, чем записей.
//
// Сессия, записанная в момент t, жива, пока now() < t+ttl. Истёкшая
// сессия не возвращается из Get, даже если фоновая чистка ещё не успела
// её удалить. Фоновая горутина раз в cleanupEvery удаляет истёкшие.
type Store struct {
	ttl time.Duration
	now func() time.Time

	mu       sync.RWMutex
	sessions map[string]session

	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

// NewStore создаёт хранилище и запускает фоновую чистку.
// now — источник времени (в тестах его подменяют).
func NewStore(ttl, cleanupEvery time.Duration, now func() time.Time) *Store {
	s := &Store{
		ttl:      ttl,
		now:      now,
		sessions: make(map[string]session),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go s.janitor(cleanupEvery)
	return s
}

func (s *Store) janitor(every time.Duration) {
	defer close(s.done)
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.cleanup()
		case <-s.stop:
			return
		}
	}
}

func (s *Store) cleanup() {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, ss := range s.sessions {
		if !now.Before(ss.expires) {
			delete(s.sessions, id) // удалять во время range по мапе можно
		}
	}
}

// Set записывает или продлевает сессию id пользователя user.
func (s *Store) Set(id, user string) {
	exp := s.now().Add(s.ttl)
	s.mu.Lock()
	s.sessions[id] = session{user: user, expires: exp}
	s.mu.Unlock()
}

// Get возвращает пользователя живой сессии.
func (s *Store) Get(id string) (string, bool) {
	now := s.now()
	s.mu.RLock() // читатели не мешают друг другу
	ss, ok := s.sessions[id]
	s.mu.RUnlock()
	// Не полагаемся на чистку: она могла ещё не дойти до этой сессии.
	if !ok || !now.Before(ss.expires) {
		return "", false
	}
	return ss.user, true
}

// Len — сколько сессий сейчас хранится (включая ещё не вычищенные).
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// Close останавливает фоновую чистку и дожидается выхода её горутины.
// Повторный Close ничего не делает.
func (s *Store) Close() {
	s.closeOnce.Do(func() { close(s.stop) }) // второй close(stop) — паника
	<-s.done
}
