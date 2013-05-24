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

// Block represents an independently calculated part of the image.
type Block struct {
	Histogram Histogram
	Picture   *Picture
	Pixels    [][]uint8

	Notify []*SubBlock // Slice of subblocks waiting for this block!
}

// PleaseNotify adds a subblock to the notification list for this block.
func (block *Block) PleaseNotify(subBlock *SubBlock) {
	block.Notify = append(block.Notify, subBlock)
}

// CalculateHistogram generates the Histogram for this block and notifies subblocks when ready.
func (block *Block) CalculateHistogram(x, y int) {

	block.Histogram = make([]int, 256)
	block.Histogram.Generate(block)
	block.Histogram.Clip(block.Picture.ClipLimit)
	block.Histogram.Map(int(block.Picture.ColorMin), int(block.Picture.ColorMax), block.Picture.BlockWidth, block.Picture.BlockHeight)

	// Notify SubBlocks waiting for this block that we're ready!
	for _, subBlock := range block.Notify {
		subBlock.WaitGroup.Done()
	}

}
