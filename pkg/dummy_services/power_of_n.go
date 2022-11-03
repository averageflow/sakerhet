package dummyservices

import "math"

// simple service that does a computation and publishes to Pub/Sub
type MyPowerOfNService interface {
	ToPowerOfN(float64, float64) float64
}

type SimplePowerOfNService struct{}

func (s SimplePowerOfNService) ToPowerOfN(x float64, n float64) float64 {
	return math.Pow(x, n)
}

func NewMyPowerOfNService() MyPowerOfNService {
	return SimplePowerOfNService{}
}
