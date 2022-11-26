package main

import (
	"image"
	"image/color"
	"strings"
	"time"
	"context"
	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

// CommandWidget is a widget displaying the output of command(s).
type RedisWidget struct {
	*BaseWidget

	keys     []string
	fonts    []string
	frames   []image.Rectangle
	colors   []color.Color
}

// NewRedisWidget returns a new RedisWidget.
func NewRedisWidget(bw *BaseWidget, opts WidgetConfig) *RedisWidget {
	bw.setInterval(time.Duration(opts.Interval)*time.Millisecond, time.Second)

	var keys, fonts, frameReps []string
	_ = ConfigValue(opts.Config["keys"], &keys)
	_ = ConfigValue(opts.Config["font"], &fonts)
	_ = ConfigValue(opts.Config["layout"], &frameReps)
	var colors []color.Color
	_ = ConfigValue(opts.Config["color"], &colors)

	layout := NewLayout(int(bw.dev.Pixels))
	frames := layout.FormatLayout(frameReps, len(keys))

	for i := 0; i < len(keys); i++ {
		if len(fonts) < i+1 {
			fonts = append(fonts, "regular")
		}
		if len(colors) < i+1 {
			colors = append(colors, DefaultColor)
		}
	}

	return &RedisWidget{
		BaseWidget: bw,
		keys:       keys,
		fonts:      fonts,
		frames:     frames,
		colors:     colors,
	}
}

// Update renders the widget.
func (w *RedisWidget) Update() error {
	size := int(w.dev.Pixels)
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for i := 0; i < len(w.keys); i++ {
		str, err := getValue(w.keys[i])
		if err != nil {
			return err
		}
		font := fontByName(w.fonts[i])

		drawString(img,
			w.frames[i],
			font,
			str,
			w.dev.DPI,
			-1,
			w.colors[i],
			image.Pt(-1, -1))
	}
	return w.render(w.dev, img)
}

func getValue(key string) (string, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    output, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return strings.TrimSuffix(string("nil"), "\n"), nil
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}
