package main
import("fmt")

type Engine struct {
	HorsePower int
}

func (e *Engine) Start(){
	fmt.Printf("Enginer started with: %d",e.HorsePower)
}

type Wheels struct{
	Count int
}

func (e *Wheels) Rotate(){
	fmt.Printf("%d wheels rotating",e.Count)
}

type Car struct{
	Engine
	Wheels
}

func main(){
	e:=Engine{80}
	w:=Wheels{4}
	car:=Car{Engine: e,Wheels: w}
	car.Start()
	car.Rotate()
	
}

