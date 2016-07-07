package main

import (
	"fmt"
	"math"
)

type Shaper interface {
	Area() float64
}

type Polygoner interface {
	Edges() int
}

type Rectangle struct {
	length, width float64
}

func (r Rectangle) Area() float64 {
	return r.length * r.width
}

func (r Rectangle) Edges() int {
	return 4
}

type Circle struct {
	radius float64
}

func (sq Circle) Area() float64 {
	return sq.radius * sq.radius * math.Pi
}

func main() {
	r := Rectangle{length: 5, width: 3}
	c := Circle{radius: 5}
	shapesArr := [...]Shaper{r, c, Circle{radius: 3}}

	fmt.Println("Looping through shapes for area and edges ...")

	for n, _ := range shapesArr {
		fmt.Printf("\nShape details: %#v\n", shapesArr[n])
		fmt.Println("Area of this shape is: ", shapesArr[n].Area())

		p, ok := shapesArr[n].(Polygoner)
		if ok {
			fmt.Printf("This shape %#v is a polygon and it has %d edges \n", p, p.Edges())
		} else {
			fmt.Printf("This shape %#v is not a polygan\n", shapesArr[n])
		}

	}

}
