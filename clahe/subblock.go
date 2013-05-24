// CLAHE implemented in Go
//
// Original CLAHE implementation by Karel Zuiderveld, karel@cv.ruu.nl
// in "Graphics Gems IV", Academic Press, 1994.
//
// Written in Go by Vic Demuzere, vic@demuzere.be
//
// License: MIT - http://opensource.org/licenses/MIT
//

package clahe

import (
	"sync"
)

// SubBlock represents a part of the image that depends on 4 Blocks.
type SubBlock struct {
	Picture *Picture

	Width  int
	Height int

	OffsetX int
	OffsetY int

	TopLeft     *Block
	TopRight    *Block
	BottomLeft  *Block
	BottomRight *Block

	WaitGroup *sync.WaitGroup
}

// Calculate waits till all dependencies are ready and interpolates this SubBlock.
func (subBlock *SubBlock) Interpolate() {

	// Wait for required blocks to complete.
	subBlock.WaitGroup.Wait()

	var inverseY, inverseX int
	var gray uint8

	// PixelCount is different for edge subblocks.
	pixelCount := subBlock.Width * subBlock.Height

	for x := 0; x < subBlock.Width; x++ {

		inverseX = subBlock.Width - x

		for y := 0; y < subBlock.Height; y++ {

			inverseY = subBlock.Height - y

			gray = subBlock.Picture.LUT[subBlock.Picture.Pixels[subBlock.OffsetY+y][subBlock.OffsetX+x]]

			// Interpolate!
			gray = uint8(((inverseY *
				((inverseX * subBlock.TopLeft.Histogram[gray]) +
					(x * subBlock.TopRight.Histogram[gray]))) +
				(y *
					((inverseX * subBlock.BottomLeft.Histogram[gray]) +
						(x * subBlock.BottomRight.Histogram[gray])))) / pixelCount)

			subBlock.Picture.Pixels[subBlock.OffsetY+y][subBlock.OffsetX+x] = gray

		}

	}

	// Notify picture that another subblock is done calculating. (yay!)
	subBlock.Picture.WaitGroup.Done()

}
