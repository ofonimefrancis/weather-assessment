package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
)

const (
	CITY_COUNT = 50
)

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

const weatherCityUrlTemplate string = "https://www.metaweather.com/api//location/search/?lattlong=%f,%f"
const weatherUrlTemplate string = "https://www.metaweather.com/api/location/%d/%d/%d/%d"
const cityUrls string = "https://public.opendatasoft.com/api/records/1.0/search/?dataset=1000-largest-us-cities-by-population-with-geographic-coordinates&facet=city&facet=state&sort=population&rows=100"

func main() {
	cityData, err := doGetRequest(cityUrls)
	if err != nil {
		panic(err)
	}

	cityDataParsed, _ := gabs.ParseJSON(cityData)
	cities, _ := cityDataParsed.Path("records").Children()
	cityCoordinates := [100]Coordinate{}
	for i, city := range cities {
		coord := city.Path("fields.coordinates").Data().([]interface{})
		cityCoordinates[i] = Coordinate{
			Latitude:  coord[0].(float64),
			Longitude: coord[1].(float64),
		}
	}
	fmt.Println(calculateAverage(cityCoordinates))
}

func calculateAverage(cityCoordinates [100]Coordinate) float64 {
	receiveChannel := make(chan float64, 50)
	sum := 0.0

	for i := 0; i < CITY_COUNT; i++ {
		go func(i int) {
			result, ok := getCurrentTemperatureForCoordinates(cityCoordinates[i])
			fmt.Println(result)
			if !ok {
				return
			}
			receiveChannel <- result
		}(i)
	}

	for i := 0; i < CITY_COUNT; i++ {
		select {
		case temp := <-receiveChannel:
			sum += temp
		}
	}
	close(receiveChannel)

	return sum / float64(CITY_COUNT)
}

func getCurrentTemperatureForCoordinates(coord Coordinate) (float64, bool) {
	weatherCityData, err := doGetRequest(fmt.Sprintf(weatherCityUrlTemplate, coord.Latitude, coord.Longitude))
	if err != nil {
		panic(err)
	}

	weatherCitiesParsed, _ := gabs.ParseJSON(weatherCityData)
	weatherCityWoeids := weatherCitiesParsed.Path("woeid").Data().([]interface{})
	weatherURLFormatted := fmt.Sprintf(weatherUrlTemplate, int64(weatherCityWoeids[0].(float64)), time.Now().Year(),
		int(time.Now().Month()), time.Now().Day())
	weatherData, err := doGetRequest(weatherURLFormatted)
	if err != nil {
		panic(err)
	}
	weatherDataParsed, _ := gabs.ParseJSON(weatherData)
	result, ok := weatherDataParsed.Path("the_temp").Data().([]interface{})[0].(float64)
	return result, ok
}

func doGetRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
