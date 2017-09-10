package main

import (
	"math"
	"testing"
	"fmt"
	"math/rand"
)

// tests configuration
const (
	maxX       = 100
	maxY       = 100
	minAP      = 3
	maxAP      = 6
	testsCount = 100
)

func randomPlainPosition() (x, y int) {
	return int(rand.Float64() * maxX), int(rand.Float64() * maxY)
}

func planePointDistance(x1, y1, x2, y2 int) (distance float64) {
	return math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2))
}

func TestCalculatePositionSeveralApd(t *testing.T) {
	// fails if GetSubscriberCoordinates returns error
	// in verbose mode prints calculation error information
	var avgErr, maxErr float64

	for testI := 0; testI < testsCount; testI++ {
		subscriberX, subscriberY := randomPlainPosition()
		apCount := int(minAP + (rand.Float64() * (maxAP - minAP)))
		testApsData := make([]AccessPointData, 0, apCount)

		for apI := 0; apI < apCount; apI ++ {
			apX, apY := randomPlainPosition()
			subscriberDistance := planePointDistance(apX, apY, subscriberX, subscriberY)
			apRssi := CalculateRSSI(subscriberDistance)
			newAp := AccessPointData{apX, apY, apRssi}
			if testing.Verbose() {
				fmt.Printf("distance: %f; calculated dis: %f; calculated rssi %d\n", subscriberDistance, newAp.GetDistanceToSubscriber(), apRssi)
			}
			testApsData = append(testApsData, newAp)
		}

		calculatedX, calculatedY, calculateErr := GetSubscriberCoordinates(testApsData)

		if calculateErr {
			t.FailNow()
		}
		if subscriberX != calculatedX || subscriberY != calculatedY {
			err := planePointDistance(subscriberX, subscriberY, calculatedX, calculatedY)
			if testing.Verbose() {
				fmt.Printf("Error amount: %v\n", err)
				fmt.Printf("\tAPD: %v\n", testApsData)
				fmt.Printf("\tDesired x y pos: %d %d\n", subscriberX, subscriberY)
				fmt.Printf("\tGot x y pos: %d %d\n", calculatedX, calculatedY)
			}
			if err > maxErr {
				maxErr = err
			}
			avgErr += err
		}
	}
	if testing.Short() || testing.Verbose() {
		fmt.Printf("Avarage error: %f\nMax error: %f\n", avgErr/testsCount, maxErr)
	}
}

func TestCalculatePositionEmptyData(t *testing.T) {
	// if apData empty, should return error
	testApsData := make([]AccessPointData, 0)
	_, _, calculateErr := GetSubscriberCoordinates(testApsData)
	if !calculateErr {
		t.Fail()
	}
}

func TestCalculatePositionTwoApd(t *testing.T) {
	// if apCount < 3 should return closest AP position
	subscriberX, subscriberY := randomPlainPosition()

	for apCount := 1; apCount < 3; apCount++ {
		testApsData := make([]AccessPointData, 0, apCount)

		for apI := 0; apI < apCount; apI ++ {
			apX, apY := randomPlainPosition()
			subscriberDistance := planePointDistance(apX, apY, subscriberX, subscriberY)
			apRssi := CalculateRSSI(subscriberDistance)
			newAp := AccessPointData{apX, apY, apRssi}
			testApsData = append(testApsData, newAp)
		}
		var closestAp *AccessPointData;
		if apCount == 2 && testApsData[1].Rssi > testApsData[0].Rssi {
			closestAp = &testApsData[1]
		} else {
			closestAp = &testApsData[0]
		}
		calculatedX, calculatedY, _ := GetSubscriberCoordinates(testApsData)
		if closestAp.X != calculatedX || closestAp.Y != calculatedY {
			t.Fail()
			break
		}
	}
}
