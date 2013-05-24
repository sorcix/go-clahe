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

type Histogram []int

// Generate calculates the base histogram for a block.
func (histogram Histogram) Generate(block *Block) {

	for x := 0; x < block.Picture.BlockWidth; x++ {
		for y := 0; y < block.Picture.BlockHeight; y++ {
			histogram[block.Picture.LUT[block.Pixels[y][x]]]++
		}
	}

}

// Clip equalizes the histogram.
func (histogram Histogram) Clip(limit int) {

	var buffer int

	binCount := len(histogram)
	excess := 0

	for i := 0; i < binCount; i++ {
		buffer = histogram[i] - limit
		if buffer > 0 {
			excess += buffer
		}
	}

	incrementPerBin := excess / binCount
	upper := binCount - incrementPerBin

	for i := 0; i < binCount; i++ {
		switch {
		case histogram[i] > limit:
			histogram[i] = limit
		case histogram[i] > upper:
			excess += upper - histogram[i]
			histogram[i] = limit
		default:
			excess -= incrementPerBin
			histogram[i] += incrementPerBin
		}
	}

	if excess > 0 {

		step := (1 + (excess / binCount))
		if step < 1 {
			step = 1
		}

		for i := 0; i < binCount; i++ {
			excess -= step
			histogram[i] += step
			if excess < 1 {
				break
			}
		}

	}

}

// Map transforms the histogram to a cumulative histogram.
func (histogram Histogram) Map(min int, max int, width int, height int) {

	// TODO: Remove floating point operations!

	sum := 0
	scale := float32(max-min)/float32(width*height)
	
	binCount := len(histogram)

	for i := 0; i < binCount; i++ {
		sum += histogram[i]
		histogram[i] = (min + int(float32(sum)*scale))
		if histogram[i] > max {
			histogram[i] = max
		}
	}
}
