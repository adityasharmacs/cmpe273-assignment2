package main

import (
  "fmt"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "time"
  "github.com/julienschmidt/httprouter"
  "net/http"
  "encoding/json"
  "strings"
)

type PositionRequest struct {
    Name    string `json:"name"`
    Address string `json:"address"`
    City    string `json:"city"`
    State   string `json:"state"`
    Zip     string `json:"zip"`
}

type PositionResponse struct {
    ID    bson.ObjectId `json:"id" bson:"_id,omitempty"`
    Name  string `json:"name"`
    Address    string `json:"address"`
    City       string `json:"city"`
    State string `json:"state"`
    Zip   string `json:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat"`
        Lng float64 `json:"lng"`
    } `json:"coordinate"`
}

type GeographicLocation struct {
    Results []struct {
        AddressComponents []struct {
            LongName  string   `json:"long_name"`
            ShortName string   `json:"short_name"`
            Types     []string `json:"types"`
        } `json:"address_components"`
        FormattedAddress string `json:"formatted_address"`
        Geometry         struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`
            } `json:"location"`
            LocationType string `json:"location_type"`
            Viewport     struct {
                Northeast struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"northeast"`
                Southwest struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"southwest"`
            } `json:"viewport"`
        } `json:"geometry"`
        PlaceID string   `json:"place_id"`
        Types   []string `json:"types"`
    } `json:"results"`
    Status string `json:"status"`
}

type GoogleLocationResponse struct {
    Results []struct {
        AddressComponents []struct {
            LongName  string   `json:"long_name"`
            ShortName string   `json:"short_name"`
            Types     []string `json:"types"`
        } `json:"address_components"`
        FormattedAddress string `json:"formatted_address"`
        Geometry         struct {
            Location struct {
                Lat float64 `json:"lat"`
                Lng float64 `json:"lng"`
            } `json:"location"`
            LocationType string `json:"location_type"`
            Viewport     struct {
                Northeast struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"northeast"`
                Southwest struct {
                    Lat float64 `json:"lat"`
                    Lng float64 `json:"lng"`
                } `json:"southwest"`
            } `json:"viewport"`
        } `json:"geometry"`
        PlaceID string   `json:"place_id"`
        Types   []string `json:"types"`
    } `json:"results"`
    Status string `json:"status"`
}

var collec *mgo.Collection
var positionResponse PositionResponse

const(
    timeout = time.Duration(time.Second*100)
)

func connectionToMongo() {
    uri := " mongodb://dbuser:dbuser@ds045464.mongolab.com:45464/testgoogledatabase"
    session, err := mgo.Dial(uri)

  	if err != nil {
    	fmt.Printf("Could not connect to mongo, go error encountered %v\n", err)
  	} else {
	  	session.SetSafe(&mgo.Safe{})
	    collec = session.DB("testgoogledatabase").C("testgoogledatabase")
	}
}



func getGoogleLoc(address string) (geoLoc GeographicLocation) {
	
	client := http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s",address)
    res, err := client.Get(url)
    if err != nil {
        fmt.Errorf("Cannot read Google API: %v", err)
    }
    defer res.Body.Close()

    decoder := json.NewDecoder(res.Body)
    err = decoder.Decode(&geoLoc)
    if(err!=nil)    {
        panic(err)
    }	
	return geoLoc
}

func getLocn(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	
	id := bson.ObjectIdHex(p.ByName("locID"))
	err := collec.FindId(id).One(&positionResponse)
  	if err != nil {
    	fmt.Printf("got an error finding a doc %v\n")
    } 	
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
    rw.WriteHeader(200)
    json.NewEncoder(rw).Encode(positionResponse)
}


func addLocn(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

    var posRequest PositionRequest
    decoder := json.NewDecoder(req.Body)
    err := decoder.Decode(&posRequest)
    if(err!=nil)    {
        fmt.Errorf("Error in decoding the Input JSON: %v", err)
    }
	address := posRequest.Address+" "+posRequest.City+" "+posRequest.State+" "+posRequest.Zip
	address = strings.Replace(address," ","%20",-1)

	locDetails := getGoogleLoc(address)

   	positionResponse.ID = bson.NewObjectId()
 	positionResponse.Address= posRequest.Address
 	positionResponse.City=posRequest.City
 	positionResponse.Name=posRequest.Name
 	positionResponse.State=posRequest.State
 	positionResponse.Zip=posRequest.Zip
	positionResponse.Coordinate.Lat=locDetails.Results[0].Geometry.Location.Lat
	positionResponse.Coordinate.Lng=locDetails.Results[0].Geometry.Location.Lng

	err = collec.Insert(positionResponse)
 	if err != nil {
   		fmt.Printf("Can't insert document: %v\n", err)
  	}

	err = collec.FindId(positionResponse.ID).One(&positionResponse)
  	if err != nil {
    	fmt.Printf("got an error finding a doc %v\n")
    } 	
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
    rw.WriteHeader(201)
    json.NewEncoder(rw).Encode(positionResponse)
}

func updateLocn(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var tempPositionResponse PositionResponse
	var positionResponse PositionResponse
	id := bson.ObjectIdHex(p.ByName("locID"))
	err := collec.FindId(id).One(&positionResponse)
  	if err != nil {
    	fmt.Printf("got an error finding a doc %v\n")
    } 
 	tempPositionResponse.Name = positionResponse.Name
 	tempPositionResponse.Address = positionResponse.Address
 	tempPositionResponse.City = positionResponse.City
 	tempPositionResponse.State = positionResponse.State
 	tempPositionResponse.Zip = positionResponse.Zip
    decoder := json.NewDecoder(req.Body)
    err = decoder.Decode(&tempPositionResponse)
	
    if(err!=nil)    {
        fmt.Errorf("Error in decoding the Input JSON: %v", err)
    }

	address := tempPositionResponse.Address+" "+tempPositionResponse.City+" "+tempPositionResponse.State+" "+tempPositionResponse.Zip
	address = strings.Replace(address," ","%20",-1)
	locationDetails := getGoogleLoc(address)
 	tempPositionResponse.Coordinate.Lat=locationDetails.Results[0].Geometry.Location.Lat
 	tempPositionResponse.Coordinate.Lng=locationDetails.Results[0].Geometry.Location.Lng
	err = collec.UpdateId(id,tempPositionResponse)
  	if err != nil {
    	fmt.Printf("Got an error while updating a Location %v\n")
    } 

	err = collec.FindId(id).One(&positionResponse)
  	if err != nil {
    	fmt.Printf("Got an error while finding a Location %v\n")
    }
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
    rw.WriteHeader(201)
    json.NewEncoder(rw).Encode(positionResponse)
}

func deleteLocn(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locID"))
	err := collec.RemoveId(id)
  	if err != nil {
    	fmt.Printf("Got an error while deleting a location %v\n")
    }
	rw.WriteHeader(200)
}

func main() {
    mux := httprouter.New()
    mux.GET("/locations/:locID", getLocn)
    mux.POST("/locations", addLocn)
    mux.PUT("/locations/:locID", updateLocn)
	mux.DELETE("/locations/:locID", deleteLocn)
    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }
	connectionToMongo()
    server.ListenAndServe()
}
