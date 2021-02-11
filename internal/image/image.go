package image

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Image struct {
	preDump    bool
	nr         int
	dumpOffset int // offset from previous dump (full)
}

func Restore(imagePath string) *Image {
	re := regexp.MustCompile("[0-9]+")
	nr, err := strconv.Atoi(re.FindString(imagePath))
	if err != nil {
		log.Error().Str("Error", err.Error()).Send()
	}

	return &Image{preDump: false, nr: nr, dumpOffset: 0}
}

func (img Image) Path() string {
	return fmt.Sprintf("%s/%s", os.Getenv("DUMP_PATH"), img.Base())
}

func (img Image) Base() string {
	prefix := 'd'
	if img.preDump {
		prefix = 'p'
	}
	return fmt.Sprintf("%c%d", prefix, img.nr)
}

func (img Image) PreDump() bool {
	return img.preDump
}

// Return the next image to dump based on this image and the dump frequency.
func (img Image) NextImage(dumpFreq int) *Image {
	return &Image{
		preDump:    img.dumpOffset < dumpFreq-1,
		nr:         img.nr + 1,
		dumpOffset: (img.nr + 1) % dumpFreq,
	}
}

// Return the next pre-dump image to dump based on this image.
func (img Image) NextPreDumpImage() *Image {
	return &Image{
		preDump:    true,
		nr:         img.nr + 1,
		dumpOffset: img.dumpOffset + 1,
	}
}

// Return the next (full) dump image to dump based on this image.
func (img Image) NextDumpImage() *Image {
	return &Image{
		preDump:    false,
		nr:         img.nr + 1,
		dumpOffset: 0,
	}
}

func FirstImage() *Image {
	return &Image{
		preDump:    false,
		nr:         0,
		dumpOffset: 0,
	}
}
