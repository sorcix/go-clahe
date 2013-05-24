package main

import (
	"flag"
	"https://github.com/sorcix/go-clahe/clahe"
)

func main() {

	input := flag.String("input", "test2.jpg", "Input file (JPG or PNG)")
	output := flag.String("output", "test2_out.jpg", "Output file (PNG)")
	bx := flag.Int("blocksX", 64, "Number of blocks (X direction)")
	by := flag.Int("blocksY", 64, "Number of blocks (Y direction)")
	limit := flag.Float64("limit", 16.0, "Relative clip limit")
	
	flag.Parse()

	picture := new(clahe.Picture)

	err := picture.Read(*input)
	
	if err != nil {
		fmt.Println(err)
	}
	
	t0 := time.Now()
	
	picture.CLAHE(*bx,*by,float32(*limit))
	
	t1 := time.Now()
	fmt.Println(t1.Sub(t0).Nanoseconds())
	
	err = picture.Write(*output)
	
	if err != nil {
		fmt.Println(err)
	}

	return

}




