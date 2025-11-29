package hn

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/patrickneise/dashboard/internal/cache"
	"github.com/patrickneise/dashboard/internal/templates"
	"github.com/patrickneise/dashboard/internal/widget"
)

type Options struct {
	Count int
	TTL   time.Duration

	Client *Client
	Log    *slog.Logger
}

func NewWidgetHandler(opts Options) http.Handler {
	count := opts.Count
	if count <= 0 {
		count = 10
	}
	if count > 50 {
		// keep it sane
		count = 50
	}

	ttl := opts.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	if opts.Client == nil {
		// Caller should supply a shared httpx-backed client, but dont' crash if not
		opts.Client = NewClient(nil)
	}

	return widget.Handler[WidgetViewModel]{
		Name:  "hn",
		TTL:   ttl,
		Cache: &cache.TTL[WidgetViewModel]{},
		Log:   opts.Log,

		Fetch: func(ctx context.Context) (WidgetViewModel, error) {
			ids, err := opts.Client.TopStories(ctx)
			if err != nil {
				var zero WidgetViewModel
				return zero, err
			}

			if len(ids) > count {
				ids = ids[:count]
			}

			items := make([]*Item, len(ids))

			// Concurrency limit so we don't open too many quesitons
			sem := make(chan struct{}, 8)

			var wg sync.WaitGroup
			errCh := make(chan error, 1)

			for i, id := range ids {
				wg.Add(1)
				go func(i int, id int64) {
					defer wg.Done()

					sem <- struct{}{}
					defer func() { <-sem }()

					it, e := opts.Client.Item(ctx, id)
					if e != nil {
						select {
						case errCh <- e:
						default:
						}
						return
					}
					items[i] = it
				}(i, id)
			}

			wg.Wait()

			select {
			case e := <-errCh:
				var zero WidgetViewModel
				return zero, e
			default:
			}

			now := time.Now()
			return BuildViewModel(now, items), nil
		},

		Render: func(vm WidgetViewModel) templ.Component {
			return HackerNewsWidgetView(vm)
		},

		Error: func(_ error) templ.Component {
			return templates.WidgetError("Hacker News", "/widgets/hn")
		},

		MarkStale: func(vm WidgetViewModel, staleBy time.Duration) WidgetViewModel {
			vm.IsStale = true
			vm.StaleBy = staleBy.Round(time.Second).String()
			return vm
		},
	}
}
