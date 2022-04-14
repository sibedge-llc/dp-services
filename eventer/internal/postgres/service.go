package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/utils"
)

type Service struct {
	ctx     context.Context
	lock    sync.Mutex
	dbs     map[uint64]*Db
	Default *Db
}

func New(ctx context.Context, cfg *config.PostgresConfig) (*Service, error) {
	s := &Service{
		ctx: ctx,
		dbs: make(map[uint64]*Db, 1),
	}

	defaultDb, err := s.Register(cfg, time.Second)
	if err != nil {
		return nil, err
	}
	s.Default = defaultDb

	return s, nil
}

func (s *Service) Register(cfg *config.PostgresConfig, timeout time.Duration) (*Db, error) {
	if cfg == nil {
		return s.Default, nil
	}

	id, err := utils.ObjectToJsonId(*cfg, false)
	if err != nil {
		return nil, fmt.Errorf("failed to make id for kafka config: %w", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	db, ok := s.dbs[id]
	if !ok {
		var err error
		if cfg.Host == "" {
			cfg.Host = s.Default.GetConfig().Host
		}
		if cfg.Db == "" {
			cfg.Db = s.Default.GetConfig().Db
		}
		if cfg.Port == 0 {
			cfg.Port = s.Default.GetConfig().Port
		}
		if cfg.User == "" {
			cfg.User = s.Default.GetConfig().User
		}
		if cfg.Password == "" {
			cfg.Password = s.Default.GetConfig().Password
		}
		if cfg.Table == "" {
			cfg.Table = s.Default.GetConfig().Table
		}
		db, err = NewDb(s.ctx, id, cfg, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to create db updater: %w", err)
		}
		s.dbs[id] = db
	}
	return db, nil
}
