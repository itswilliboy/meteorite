package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type MediaMetadata struct {
	Width      int
	Height     int
	DurationMs int
	Bitrate    int
	Codec      string
	Framerate  float64
	SampleRate int
	Channels   int
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

type ffprobeDisposition struct {
	AttachedPic int `json:"attached_pic"`
}

type ffprobeStream struct {
	CodecType   string             `json:"codec_type"`
	CodecName   string             `json:"codec_name"`
	Width       int                `json:"width"`
	Height      int                `json:"height"`
	SampleRate  string             `json:"sample_rate"`
	Channels    int                `json:"channels"`
	RFrameRate  string             `json:"r_frame_rate"`
	BitRate     string             `json:"bit_rate"`
	Disposition ffprobeDisposition `json:"disposition"`
}

type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

func parseFrameRate(s string) float64 {
	num, den, ok := strings.Cut(s, "/")
	if !ok {
		return 0
	}

	n, err1 := strconv.ParseFloat(num, 64)
	d, err2 := strconv.ParseFloat(den, 64)
	if err1 != nil || err2 != nil || d == 0 {
		return 0
	}

	return n / d
}

func ProbeMedia(data []byte, ext string) (*MediaMetadata, error) {
	tmp, err := os.CreateTemp("", "probe-*"+ext)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return nil, err
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_format", "-show_streams", tmp.Name())

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		return nil, err
	}

	meta := &MediaMetadata{}

	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		meta.DurationMs = int(d * 1000)
	}
	if b, err := strconv.Atoi(probe.Format.BitRate); err == nil {
		meta.Bitrate = b
	}

	for _, s := range probe.Streams {
		// skip embedded cover-art picture streams (e.g. MP3/FLAC cover art),
		// they're tagged as a "video" stream but aren't real video content
		if s.Disposition.AttachedPic != 0 {
			continue
		}

		switch s.CodecType {
		case "video":
			meta.Width = s.Width
			meta.Height = s.Height
			meta.Codec = s.CodecName
			if fr := parseFrameRate(s.RFrameRate); fr > 0 {
				meta.Framerate = fr
			}
		case "audio":
			if meta.Codec == "" {
				meta.Codec = s.CodecName
			}
			if sr, err := strconv.Atoi(s.SampleRate); err == nil {
				meta.SampleRate = sr
			}
			meta.Channels = s.Channels
		}
	}

	return meta, nil
}

// pulls embedded album art out of an audio file, if present
func ExtractCoverArt(data []byte, ext string) ([]byte, error) {
	tmp, err := os.CreateTemp("", "cover-*"+ext)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return nil, err
	}

	cmd := exec.Command("ffmpeg", "-v", "error", "-i", tmp.Name(), "-an", "-c:v", "copy", "-f", "image2pipe", "-")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// no embedded video/picture stream
		return nil, nil
	}

	if out.Len() == 0 {
		return nil, nil
	}

	return out.Bytes(), nil
}
