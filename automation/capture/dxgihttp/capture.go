package dxgihttp

import (
	"context"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Nigh/autodream/automation/capture"
	"github.com/Nigh/autodream/automation/frame"
	"github.com/Nigh/autodream/automation/internal/imgutil"
	"github.com/Nigh/autodream/automation/log"
)

// Config configures the Nigh/dxgi-capture HTTP client.
//
// ponytail: PNG+HTTP per frame (decode + full buffer alloc). Upgrade path = capture/dxgi.
type Config struct {
	BaseURL   string // e.g. http://127.0.0.1:3000
	DisplayID int
	// Optional region crop on the server. Zero W/H means full capture.
	Region frame.ROI
	Client *http.Client
	Logger log.Logger
}

// Capture implements capture.Capture against the dxgi-capture HTTP API.
type Capture struct {
	cfg     Config
	client  *http.Client
	log     log.Logger
	pool    frame.Pool
	started atomic.Bool
	mu      sync.Mutex
	current *frame.Frame
}

// New builds an HTTP capture client.
func New(cfg Config) (*Capture, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("dxgihttp: empty BaseURL")
	}
	if _, err := url.Parse(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("dxgihttp: BaseURL: %w", err)
	}
	c := cfg.Client
	if c == nil {
		c = &http.Client{Timeout: 5 * time.Second}
	}
	lg := cfg.Logger
	if lg == nil {
		lg = log.Nop()
	}
	return &Capture{cfg: cfg, client: c, log: lg}, nil
}

func (c *Capture) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.started.Store(true)
	return nil
}

func (c *Capture) Stop() error {
	c.started.Store(false)
	c.mu.Lock()
	old := c.current
	c.current = nil
	c.mu.Unlock()
	if old != nil {
		c.pool.Release(old)
	}
	return nil
}

func (c *Capture) CurrentFrame() (*frame.Frame, error) {
	if !c.started.Load() {
		return nil, capture.ErrNotStarted
	}
	f, err := c.fetch(context.Background())
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	old := c.current
	c.current = f
	c.mu.Unlock()
	if old != nil {
		c.pool.Release(old)
	}
	return f, nil
}

func (c *Capture) fetch(ctx context.Context) (*frame.Frame, error) {
	u, err := c.captureURL()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("dxgihttp: request: %w", err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dxgihttp: get: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return nil, fmt.Errorf("dxgihttp: status %d: %s", res.StatusCode, body)
	}
	img, err := png.Decode(res.Body)
	if err != nil {
		return nil, fmt.Errorf("dxgihttp: png: %w", err)
	}
	w, h, data := imgutil.ImageToBGRA(img)
	f := c.pool.Acquire(w, h, frame.PixelFormatBGRA8)
	copy(f.Data, data)
	c.log.Trace("dxgihttp frame", "id", f.FrameID, "w", w, "h", h)
	return f, nil
}

func (c *Capture) captureURL() (string, error) {
	base, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return "", fmt.Errorf("dxgihttp: BaseURL: %w", err)
	}
	q := url.Values{}
	q.Set("display_id", strconv.Itoa(c.cfg.DisplayID))
	r := c.cfg.Region
	var path string
	if r.W > 0 && r.H > 0 {
		path = "/capture/region"
		q.Set("x", strconv.Itoa(r.X))
		q.Set("y", strconv.Itoa(r.Y))
		q.Set("width", strconv.Itoa(r.W))
		q.Set("height", strconv.Itoa(r.H))
	} else {
		path = "/capture/full"
	}
	rel := &url.URL{Path: path, RawQuery: q.Encode()}
	return base.ResolveReference(rel).String(), nil
}
