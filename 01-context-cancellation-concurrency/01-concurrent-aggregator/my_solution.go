package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"bytes"

	"github.com/fithanso/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/order"
	"github.com/fithanso/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/profile"

	"golang.org/x/sync/errgroup"
)

type Option func (*UserAggregator)


func WithLogger(l *slog.Logger) Option{
	return func(ua *UserAggregator) {
		ua.logger = l
	}
}

func WithTimeout(to time.Duration) Option{
	return func(ua *UserAggregator) {
		ua.timeout = to
	}
}

type AggregatedProfile struct {
	Name    string
	Cost    float64
}


type UserAggregator struct {
	orderService   order.Service
	profileService profile.Service
	timeout        time.Duration
	logger         *slog.Logger
}

func NewUserAggregator(os order.Service, ps profile.Service, opts ...Option) *UserAggregator {
	ua := &UserAggregator{
		orderService: os,
		profileService: ps,
		logger: slog.Default(),
	}
	
	for _, opt := range opts {
		opt(ua)
	}
	
	return ua
}

func (ua *UserAggregator) Aggregate(ctx context.Context, id int) ([]*AggregatedProfile, error) {
	
	var (
		rootCtx  context.Context
		cancel   context.CancelFunc
		ors      []*order.Order
		prof     *profile.Profile
	)
	
	if ua.timeout > 0 {
		rootCtx, cancel = context.WithTimeout(ctx, ua.timeout)
	}else{
		rootCtx, cancel = context.WithCancel(ctx)
	}
	
	defer cancel()
	
	ua.logger.Info("starting aggregation", "user_id", id)
	group, gCtx := errgroup.WithContext(rootCtx)
	
	// Call order service
	group.Go(func() error {
		orders, err := ua.orderService.GetAll(gCtx, id)
		if err != nil {
			ua.logger.Warn("failed to fetch orders", "user_id", id, "err", err)
			return fmt.Errorf("orders fetch failed: %w", err)
		}
		ors = orders
		
		return nil
	})
	
	// Call profile service
	group.Go(func() error {
		profile, err := ua.profileService.Get(gCtx, id)
		if err != nil {
			ua.logger.Warn("failed to fetch profile info", "user_id", id, "err", err)
			return fmt.Errorf("profile fetch failed: %w", err)
		}
		prof = profile
		return nil
	})
	
	err := group.Wait()
	if err != nil {
		ua.logger.Warn("aggregator exited with error", "user_id", id, "err", err)
		return nil, err
	}
	
	var aps []*AggregatedProfile
	
	for _, o := range ors {
		if o.UserId == prof.Id {
			aps = append(aps, &AggregatedProfile{Name: prof.Name, Cost: o.Cost})
		}
	}
	
	ua.logger.Info("aggregation complete successfully", "user_id", id, "count", len(aps))
	return aps, nil
	
}



func main() {
	basicProfiles := []*profile.Profile{
		{Id: 1, Name: "Alice"},
		{Id: 2, Name: "Bob"},
		{Id: 3, Name: "Charlie"},
		{Id: 4, Name: "Dave"},
		{Id: 5, Name: "Eva"},
	}
	
	psm := &ProfileServiceMock{
		2 * time.Second,
		false,
		basicProfiles,
	}
	osm := &OrderServiceMock{
		20 * time.Second,
		false,
		[]*order.Order{
			{1, 1, 100.0},
			{2, 1, 20.6},
			{3, 3, 30.79},
		},
	}
	
	var buf bytes.Buffer
	spyLogger := slog.New(slog.NewJSONHandler(&buf, nil))
	au := NewUserAggregator(osm, psm, WithLogger(spyLogger), WithTimeout(3 * time.Second))
	
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	
	aggregatedProfiles, err := au.Aggregate(ctx, 1)
	fmt.Println(aggregatedProfiles)
	fmt.Println(err)
}