package main

import (
	"os"
	"fmt"
	"math"
	"errors"
	"strings"
	"encoding/json"
)

type AccessPointData struct {
	X    int `json:"x"`
	Y    int `json:"y"`
	Rssi int `json:"rssi"`
}

type point struct {
	X, Y float64
}

func main() {
	var err error
	var apdArr []AccessPointData

	// get executable name from arg
	name := os.Args[0]
	pathTokens := []string{"\\", "/"}
	for _, token := range pathTokens {
		items := strings.Split(name, token)
		name = items[len(items)-1]
	}

	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		err = json.Unmarshal([]byte(firstArg), &apdArr)
	} else {
		err = errors.New("Argument required")
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		usageInfo := "Accepts single argument - json array of {\"x\": int, \"y\": int, \"rssi\": int}\n" +
			"Usage: %s \"[{\\\"x\\\": 0,\\\"y\\\": 0,\\\"rssi\\\": -50}, {\\\"x\\\": 10,\\\"y\\\": 10,\\\"rssi\\\": -60}, {\\\"x\\\": 30,\\\"y\\\": 40,\\\"rssi\\\": -80}]\""
		fmt.Printf(usageInfo, name)
	} else {
		x, y, err := GetSubscriberCoordinates(apdArr)
		if err {
			fmt.Printf("Error while calculating subscriber position")
		} else {
			fmt.Printf("Subscriber coordinates x:%d y:%d", x, y)
		}
	}
}

func (ap AccessPointData) toPoint() point {
	return point{float64(ap.X), float64(ap.Y)}
}

func (p1 point) distanceTo(p2 point) (distance float64) {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
}

func circleIntersection(c1p point, c1r float64, c2p point, c2r float64) (p1, p2 point, err bool) {
	// http://stackoverflow.com/a/3349134/798588
	dx := c2p.X - c1p.X
	dy := c2p.Y - c1p.Y
	d := math.Sqrt(dx*dx + dy*dy)
	if d > c1r+c2r {
		// no solutions, increase circe radius with little overlap
		rD := (d - (c1r + c2r)) * 0.6
		c1r += rD
		c2r += rD
	}
	if d < math.Abs(c1r-c2r) {
		// no solutions, decrease circe radius with little overlap
		rD := (math.Abs(c1r-c2r) - d) * 0.6
		if c1r > c2r {
			c1r -= rD
			c2r += rD
		} else {
			c1r += rD
			c2r -= rD
		}
	}
	if d == 0 && c1r == c2r {
		// the circles are coincident and there are an infinite number of solutions.
		err = true
		return
	}
	a := (c1r*c1r - c2r*c2r + d*d) / (2 * d)
	h := math.Sqrt(c1r*c1r - a*a)
	xm := c1p.X + a*dx/d
	ym := c1p.Y + a*dy/d
	p1.X = xm + h*dy/d
	p1.Y = ym - h*dx/d
	p2.X = xm - h*dy/d
	p2.Y = ym + h*dx/d
	return
}

func GetSubscriberCoordinates(accessPoints []AccessPointData) (x, y int, err bool) {
	if len(accessPoints) == 0 {
		// if apData empty return error
		err = true
		return
	} else if len(accessPoints) < 3 {
		// if apCount < 3 return closest AP position
		if len(accessPoints) == 2 && accessPoints[1].Rssi > accessPoints[0].Rssi {
			x = accessPoints[1].X
			y = accessPoints[1].Y
		} else {
			x = accessPoints[0].X
			y = accessPoints[0].Y
		}
		return
	} else {
		// calculate subscriber position
		var p1, p2 point
		var ap1, ap2 *AccessPointData

		// find intersection of 1st and 2nd AP
		ap1 = &accessPoints[0]
		ap2 = &accessPoints[1]
		p1, p2, _ = circleIntersection(ap1.toPoint(), ap1.GetDistanceToSubscriber(), ap2.toPoint(), ap2.GetDistanceToSubscriber())

		// for each combination of APs
		// combine pairs of closest points
		var p1d, p2d point
		deltaCount := 1
		for i1 := 0; i1+1 < len(accessPoints); i1++ {
			ap1 = &accessPoints[i1]
			for i2 := i1 + 1; i2 < len(accessPoints); i2++ {
				if i1 == 0 && i2 == 1 {
					// intersection of accessPoints[0] and accessPoints[1] already found
					continue
				}
				ap2 = &accessPoints[i2]
				p3, p4, _ := circleIntersection(ap1.toPoint(), ap1.GetDistanceToSubscriber(), ap2.toPoint(), ap2.GetDistanceToSubscriber())

				// make p1 closest to one of new points
				if math.Min(p1.distanceTo(p3), p1.distanceTo(p4)) > math.Min(p2.distanceTo(p3), p2.distanceTo(p4)) {
					p1, p2 = p2, p1
					p1d, p2d = p2d, p1d
				}
				// make p3 closest to one of old points
				if p1.distanceTo(p3) > p1.distanceTo(p4) {
					p3, p4 = p4, p3
				}
				// combine p1 with p3, p2 with p4
				p1d.X += p3.X - p1.X
				p1d.Y += p3.Y - p1.Y
				p2d.X += p4.X - p2.X
				p2d.Y += p4.Y - p2.Y
				deltaCount++
			}
		}

		// make p1 point within point-dense area
		var zero point
		if zero.distanceTo(p1d) > zero.distanceTo(p2d) {
			p1, p2 = p2, p1
			p1d, p2d = p2d, p1d
		}
		// + .5 for rounding purpose
		x = int(p1.X + (p1d.X / float64(deltaCount)) + .5)
		y = int(p1.Y + (p1d.Y / float64(deltaCount)) + .5)
		return
	}
}

func CalculateRSSI(distance float64) (rssi int) {
	fq := 2400.0
	N := 27.0
	if distance < 1 {
		distance = 1.0
	}
	rssi = -int(20*math.Log10(fq) + N*math.Log10(distance) - 28.)
	return
}

func (ap *AccessPointData) GetDistanceToSubscriber() (distance float64) {
	fq := 2400.0
	N := 27.0
	rssi := float64(ap.Rssi)
	distance = math.Pow(10, (28.0-rssi-20*math.Log10(fq))/N)
	return
}
