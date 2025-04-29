package model

import (
	"math"
)

type Bearing struct {
	angle       float64
	declination float64
}

func NewBearing(declination float64) *Bearing {
	bearing := &Bearing{
		declination: declination,
	}

	return bearing
}

// update bearing with new sensor data
func (b *Bearing) SetInt(x, y int32) {
	b.angle = math.Atan2(float64(y), float64(x)) + b.declination
}

func (b *Bearing) SetFloat(x, y float64) {
	b.angle = math.Atan2(y, x) + b.declination
}

// bearing angle from North in radians
func (b *Bearing) Angle() float64 {
	return b.angle
}

// bearing angle from North in degrees
func (b *Bearing) AngleDeg() float64 {
	return b.angle * (180 / math.Pi)
}
