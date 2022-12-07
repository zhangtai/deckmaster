package main

import (
	"context"
	"image"
	"image/color"
	"log"
	"strings"
	"time"
)

var ctx = context.Background()

// CommandWidget is a widget displaying the output of command(s).
type RedisWidget struct {
	*BaseWidget

	keys      []string
	fonts     []string
	icon_path string
	icon_key  string
	icon      image.Image
	frames    []image.Rectangle
	colors    []color.Color
}

// NewRedisWidget returns a new RedisWidget.
func NewRedisWidget(bw *BaseWidget, opts WidgetConfig) *RedisWidget {
	bw.setInterval(time.Duration(opts.Interval)*time.Millisecond, time.Second)

	var keys, fonts, frameReps []string
	var icon_key string
	_ = ConfigValue(opts.Config["keys"], &keys)
	_ = ConfigValue(opts.Config["font"], &fonts)
	_ = ConfigValue(opts.Config["icon_key"], &icon_key)
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

	w := &RedisWidget{
		BaseWidget: bw,
		keys:       keys,
		fonts:      fonts,
		icon_key:   icon_key,
		frames:     frames,
		colors:     colors,
	}
	if icon_key != "" {
		log.Println("Reading icon image")
		icon_path, _, err := getValue(icon_key)
		if err != nil {
			log.Printf("Failed to read icon_path from icon_key: %s", icon_key)
			return nil
		}
		if err := w.LoadImage(icon_path); err != nil {
			log.Printf("Failed LoadImage: %s", icon_path)
			return nil
		}
	}
	return w
}

func (w *RedisWidget) LoadImage(path string) error {
	path, err := expandPath(w.base, path)
	if err != nil {
		return err
	}
	icon, err := loadImage(path)
	if err != nil {
		size := int(w.dev.Pixels)
		w.icon = image.NewRGBA(image.Rect(0, 0, size, size))
	} else {
		w.icon = icon
	}

	return nil
}

// Update renders the widget.
func (w *RedisWidget) Update() error {
	size := int(w.dev.Pixels)
	margin := size / 18
	height := size - (margin * 2)
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for i := 0; i < len(w.keys); i++ {
		str, color, err := getValue(w.keys[i])
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
			color,
			image.Pt(-1, -1))
	}

	if w.icon_key != "" {
		icon_path, _, err := getValue(w.icon_key)
		if err != nil {
			return nil
		}

		if w.icon_path != icon_path {
			if err := w.LoadImage(icon_path); err != nil {
				log.Printf("Failed LoadImage: %s", icon_path)
				return nil
			}
			log.Printf("Setting icon_path to: %s", icon_path)
			w.icon_path = icon_path
		}

		drawErr := drawImage(img,
			w.icon,
			height,
			image.Pt(-1, -1))

		if drawErr != nil {
			return drawErr
		}
	}
	return w.render(w.dev, img)
}

func getValue(key string) (string, color.Color, error) {
	output, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return strings.TrimSuffix(string("nil"), "\n"), DefaultColor, nil
	}
	return strings.TrimSuffix(string(output), "\n"), DefaultColor, nil
}
