package main

import(
	"fmt"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct{
	Radius float64
}

func (c *Circle) Area() float64 {
	return 2*3.14*c.Radius
}

func (c *Circle) Perimeter() float64 {
	return 2*c.Radius
}

type Rectangle struct{
	Width float64
	Height float64
}


func (c *Rectangle) Area() float64 {
	return c.Height*c.Width
}

func (c *Rectangle) Perimeter() float64 {
	return 2*(c.Height+c.Width)
}


type Triangle struct{
	Base float64
	Height float64
	SideA float64
	SideB float64
	SideC float64
	
}


func (c Triangle) Area() float64 {
	return 0.5*c.Base*c.Height
}

func (c *Triangle) Perimeter() float64 {
	return c.SideA+c.SideB+c.SideC
}

func PrintShapeInfo(s Shape){
	fmt.Printf("Area: %.2f, Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

func main(){
	c:=Circle{2}
	r:=Rectangle{4,5}
	t:=Triangle{3,4,3,4,2}
	ShapeSlice:=make([]Shape,3)
	ShapeSlice[0]=&c
	ShapeSlice[1]=&r
	ShapeSlice[2]=&t
	for _,s:=range(ShapeSlice){
		PrintShapeInfo(s)
	}
	
	
}
