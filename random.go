package main

import (
	"math"
	"math/rand"
	"reflect"
	"time"
)

func randomBetween(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

func randomElement(list interface{}) reflect.Value {
	rand.Seed(time.Now().UTC().UnixNano())
	rnd := rand.Float64()
	return randomElementWithRand(list, rnd)
}

func randomElementWithRand(list interface{}, rnd float64) reflect.Value {
	listVal := reflect.ValueOf(list)
	index := int(math.Floor(rnd * (float64(listVal.Len()) - 1)))
	return listVal.Index(index)
}

func randomGaussian(variance float64) float64 {
	var r float64
	for i := variance; i > 0; i-- {
		rand.Seed(time.Now().UTC().UnixNano())
		r += rand.Float64()
	}
	return r / variance
}

func transformValue(rnd float64, threshold float64) float64 {
	offset := 1 - threshold
	val := rnd - threshold
	normalized := (100.0 / offset) * val
	var value float64
	if normalized <= 0 {
		value = 0
	} else {
		value = 0.01 * normalized
	}
	return value
}

func gaussianRandomElement(list interface{}, variance float64, threshold float64) reflect.Value {
	rnd := randomGaussian(variance)
	value := transformValue(math.Abs(rnd-0.5)*2, threshold)
	return randomElementWithRand(list, value)
}
