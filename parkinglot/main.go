package main

import (
	"errors"
	"fmt"
	"time"
)

type Vehicle interface {
	Type() string
}

type Car struct {
	id   string
	cost int
}

func (c *Car) Type() string {
	return "car"
}

type Motorcycle struct {
	id   string
	cost int
}

func (c *Motorcycle) Type() string {
	return "motorcycle"
}

type Truck struct {
	id   string
	cost int
}

func (c *Truck) Type() string {
	return "truck"
}

type Spot struct {
	typeof   string
	cost     int
	occupied bool
}

type Floor struct {
	Number int
	Spots  []Spot
}

type ParkingLot struct {
	Floors []Floor
}

type Ticket struct {
	Vehicle   Vehicle
	Spot      *Spot
	EntryTime time.Time
	Unparked  bool
}

func (p *ParkingLot) findParkingSpot(v Vehicle) (*Spot, error) {
	typi := v.Type()

	for i := range p.Floors {
		for j := range p.Floors[i].Spots {
			spot := &p.Floors[i].Spots[j]

			if spot.typeof == typi && !spot.occupied {
				return spot, nil
			}
		}
	}

	return nil, errors.New("parking lot full for this vehicle type")
}

// availability returns the number of free spots for each vehicle type.
func (p *ParkingLot) availability() map[string]int {
	available := map[string]int{
		"car":        0,
		"motorcycle": 0,
		"truck":      0,
	}

	for _, floor := range p.Floors {
		for _, spot := range floor.Spots {
			if !spot.occupied {
				available[spot.typeof]++
			}
		}
	}

	return available
}

// park parks a vehicle and returns its ticket.
func (p *ParkingLot) park(v Vehicle) (*Ticket, error) {
	spot, err := p.findParkingSpot(v)
	if err != nil {
		return nil, err
	}

	spot.occupied = true

	ticket := &Ticket{
		Vehicle:   v,
		Spot:      spot,
		EntryTime: time.Now(),
		Unparked:  false,
	}

	return ticket, nil
}

// unpark removes the vehicle from the parking spot
// and calculates the parking cost.
func (p *ParkingLot) unpark(t *Ticket) (int, error) {
	if t == nil || t.Spot == nil {
		return 0, errors.New("invalid ticket")
	}

	if t.Unparked {
		return 0, errors.New("vehicle already unparked")
	}

	elapsed := time.Since(t.EntryTime)

	// Round up to the next hour.
	hours := (elapsed + time.Hour - 1) / time.Hour

	cost := int(hours) * t.Spot.cost

	t.Spot.occupied = false
	t.Unparked = true

	return cost, nil
}

func main() {

	// Create parking spots.
	parkingLot := ParkingLot{
		Floors: []Floor{
			{
				Number: 1,
				Spots: []Spot{
					{typeof: "car", cost: 50},
					{typeof: "car", cost: 50},
					{typeof: "motorcycle", cost: 30},
				},
			},
			{
				Number: 2,
				Spots: []Spot{
					{typeof: "truck", cost: 100},
					{typeof: "truck", cost: 100},
					{typeof: "car", cost: 50},
				},
			},
		},
	}

	// Check availability before parking.
	fmt.Println("Availability before parking:")
	fmt.Println(parkingLot.availability())

	// Create vehicles.
	car := &Car{
		id:   "CAR-101",
		cost: 50,
	}

	motorcycle := &Motorcycle{
		id:   "BIKE-101",
		cost: 30,
	}

	truck := &Truck{
		id:   "TRUCK-101",
		cost: 100,
	}

	// Park vehicles.
	carTicket, err := parkingLot.park(car)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	bikeTicket, err := parkingLot.park(motorcycle)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	truckTicket, err := parkingLot.park(truck)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nAvailability after parking:")
	fmt.Println(parkingLot.availability())

	fmt.Println("\nVehicles parked successfully.")

	// Just for demonstration, pretend some time has passed.
	// In a real program EntryTime would naturally be in the past.
	carTicket.EntryTime = time.Now().Add(-2*time.Hour - 15*time.Minute)
	bikeTicket.EntryTime = time.Now().Add(-45 * time.Minute)
	truckTicket.EntryTime = time.Now().Add(-3*time.Hour - 10*time.Minute)

	// Unpark vehicles.
	carCost, err := parkingLot.unpark(carTicket)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	bikeCost, err := parkingLot.unpark(bikeTicket)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	truckCost, err := parkingLot.unpark(truckTicket)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\nParking charges:")
	fmt.Println("Car:", carCost)
	fmt.Println("Motorcycle:", bikeCost)
	fmt.Println("Truck:", truckCost)

	fmt.Println("\nAvailability after unparking:")
	fmt.Println(parkingLot.availability())
}