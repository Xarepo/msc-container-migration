package image

import (
	"fmt"
	"io/ioutil"
	"math"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/Xarepo/msc-container-migration/internal/env"
)

// imageType represents the type of the dump
//
// preDump:
// This dump only contains the memory changed since last dump. It is not
// possiblle to restore from dumps of this type.
//
// fullDump:
// This dump contains the full state of the application. It is possible to
// restore from dumps of this type.
//
// checkpoint:
// This dump is the same as fullDump, but it has been issued manually via an
// IPC rather than automatically by the system. Note that checkpoint images may
// have the same number as another fulldump or preDump image.
type imageType int

const (
	preDump imageType = iota
	fullDump
	checkpoint
)

type Image struct {
	_type imageType
	nr    int

	// Offset from previous dump (full). Will always be zero for full dumps.
	dumpOffset int
}

// Construct an image based on a imagePath.
// Used when restoring containers.
func Restore(imagePath string) *Image {
	re := regexp.MustCompile("[0-9]+")
	nr, err := strconv.Atoi(re.FindString(imagePath))
	if err != nil {
		log.Error().Str("Error", err.Error()).Send()
	}

	return &Image{_type: fullDump, nr: nr, dumpOffset: 0}
}

// Construct a checkpoint image from another image.
func (img *Image) Checkpoint() *Image {
	return &Image{
		_type:      checkpoint,
		dumpOffset: 0,
		nr:         img.nr + 1,
	}
}

// Retrieve the latest dump possible to recover from, i.e. the last transferred
// full dump.
//
// The latest possible dump to recover from is determined by finding the
// directory fullfilling:
// 1) its name starts with the character 'd', i.e. the name is of the form
// "dX" (where X is an integer)
// 2) It has the highest number in its name of all full dumps (directories
// prefixed with 'd'), i.e. for all directories of the form "dX", its X is
// maximal.
func Recover() *Image {
	entries, err := ioutil.ReadDir(env.Getenv().DUMP_PATH)
	if err != nil {
		log.Error().
			Str("Error", err.Error()).
			Msg("Failed to read recovery directory")
	}

	// Find the directory that fullfills the requirements described in the
	// function comment.
	latestDumpNum := math.Inf(-1)
	r := regexp.MustCompile("[0-9]+")
	for _, entry := range entries {
		num, _ := strconv.Atoi(r.FindString(entry.Name()))
		if entry.IsDir() && entry.Name()[0] == 'd' && float64(num) > latestDumpNum {
			latestDumpNum = float64(num)
		}
	}

	latestDump := fmt.Sprintf("d%d", int(latestDumpNum))
	log.Debug().
		Str("Dump", latestDump).
		Msg("Latest possible recovery dump determined")
	return &Image{
		_type: fullDump, nr: int(latestDumpNum), dumpOffset: 0,
	}
}

func (img Image) Path() string {
	return fmt.Sprintf("%s/%s", env.Getenv().DUMP_PATH, img.Base())
}

func (img Image) Base() string {
	prefix := '_'
	switch img._type {
	case preDump:
		prefix = 'p'
	case fullDump:
		prefix = 'd'
	case checkpoint:
		prefix = 'c'
	default:
		log.Panic().Int("Type", int(img._type)).Msg("Image has invalid type")
	}
	return fmt.Sprintf("%c%d", prefix, img.nr)
}

// Return whether of not the image is a predump
func (img Image) PreDump() bool {
	return img._type == preDump
}

// Return the next image to dump based on this image and the dump frequency.
func (img Image) NextImage(dumpFreq int) *Image {
	t := fullDump
	if img.dumpOffset < dumpFreq-1 {
		t = preDump
	}
	return &Image{
		_type:      t,
		nr:         img.nr + 1,
		dumpOffset: (img.nr + 1) % dumpFreq,
	}
}

// Return the next pre-dump image to dump based on this image.
func (img Image) NextPreDumpImage() *Image {
	return &Image{
		_type:      preDump,
		nr:         img.nr + 1,
		dumpOffset: img.dumpOffset + 1,
	}
}

// Return the next (full) dump image to dump based on this image.
func (img Image) NextDumpImage() *Image {
	return &Image{
		_type:      fullDump,
		nr:         img.nr + 1,
		dumpOffset: 0,
	}
}

// Return the first of all images, across all hosts.
func FirstImage() *Image {
	return &Image{
		_type:      fullDump,
		nr:         0,
		dumpOffset: 0,
	}
}
