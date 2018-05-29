package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	apiURL         = "https://api.nal.usda.gov/ndb"
	searchEndpoint = "/search/"
	listEndpoint   = "/list/"
	ndbnoEndpoint  = "/V2/reports"
)

// USDA is an interface to hold common methods from the stuff returned by USDA API
type USDA interface {
	String() string
}

// FoodByNDBO is a struct to hold JSON returned by the Ndbo search endpoint
type FoodByNDBO struct {
	Foods []struct {
		Food struct {
			Sr   string `json:"sr"`
			Type string `json:"type"`
			Desc struct {
				Ndbno string  `json:"ndbno"`
				Name  string  `json:"name"`
				Sd    string  `json:"sd"`
				Fg    string  `json:"fg"`
				Sn    string  `json:"sn"`
				Cn    string  `json:"cn"`
				Manu  string  `json:"manu"`
				Nf    float64 `json:"nf"`
				Cf    float64 `json:"cf"`
				Ff    float64 `json:"ff"`
				Pf    float64 `json:"pf"`
				R     string  `json:"r"`
				Rd    string  `json:"rd"`
				Ds    string  `json:"ds"`
				Ru    string  `json:"ru"`
			} `json:"desc"`
			Nutrients []struct {
				NutrientID string `json:"nutrient_id"`
				Name       string `json:"name"`
				Group      string `json:"group"`
				Unit       string `json:"unit"`
				Value      string `json:"value"`
				Derivation string `json:"derivation"`
				Sourcecode string `json:"sourcecode"`
				Dp         int    `json:"dp"`
				Se         string `json:"se"`
				Measures   []struct {
					Label string  `json:"label"`
					Eqv   float64 `json:"eqv"`
					Eunit string  `json:"eunit"`
					Qty   float64 `json:"qty"`
					Value string  `json:"value"`
				} `json:"measures"`
			} `json:"nutrients"`
			Sources []struct {
				ID      int    `json:"id"`
				Title   string `json:"title"`
				Authors string `json:"authors"`
				Vol     string `json:"vol"`
				Iss     string `json:"iss"`
				Year    string `json:"year"`
			} `json:"sources"`
			Footnotes []interface{} `json:"footnotes"`
			Langual   []interface{} `json:"langual"`
		} `json:"food"`
	} `json:"foods"`
	Count    int     `json:"count"`
	Notfound int     `json:"notfound"`
	API      float64 `json:"api"`
}

// FoodToStore is the struct that will be saved as a json to a json file to save calorie intake
type FoodToStore struct {
	Name         string
	Energy       float64
	Protein      float64
	Fat          float64
	Carbohydrate float64
	Fiber        float64
	Qtd          float64
}

// SearchList is a struct to hold JSON returned by the search endpoint
type SearchList struct {
	List struct {
		Q     string `json:"q"`
		Sr    string `json:"sr"`
		Ds    string `json:"ds"`
		Start int    `json:"start"`
		End   int    `json:"end"`
		Total int    `json:"total"`
		Group string `json:"group"`
		Sort  string `json:"sort"`
		Item  []struct {
			Offset int    `json:"offset"`
			Group  string `json:"group"`
			Name   string `json:"name"`
			Ndbno  string `json:"ndbno"`
			Ds     string `json:"ds"`
			Manu   string `json:"manu"`
		} `json:"item"`
	} `json:"list"`
}

// FoodList is a struct to hold JSON returned by the list endpoint
type FoodList struct {
	List struct {
		Lt    string `json:"lt"`
		Start int    `json:"start"`
		End   int    `json:"end"`
		Total int    `json:"total"`
		Sr    string `json:"sr"`
		Sort  string `json:"sort"`
		Item  []struct {
			Offset int    `json:"offset"`
			ID     string `json:"id"`
			Name   string `json:"name"`
		} `json:"item"`
	} `json:"list"`
}

func (sl SearchList) String() string {
	var res string
	for _, item := range sl.List.Item {
		res += fmt.Sprintf("%d: {\n  Group: %s\n  Name: %s\n  Ndbno: %s\n  Database: %s\n  Manufacturer: %s\n}\n",
			item.Offset, item.Group, item.Name, item.Ndbno, item.Ds, item.Manu)
	}
	return res
}

func (fl FoodList) String() string {
	var res string
	for _, item := range fl.List.Item {
		res += fmt.Sprintf("%d: {\n  Ndbno: %s\n  Name: %s\n}\n",
			item.Offset, item.ID, item.Name)
	}
	return res
}

func (fndbo FoodByNDBO) String() string {
	var res string
	for _, food := range fndbo.Foods {
		res += fmt.Sprintf("{\n  Ndbno: %s\n  Name: %s\n  Nutrients: [\n",
			food.Food.Desc.Ndbno, food.Food.Desc.Name)
		for _, nutrient := range food.Food.Nutrients {
			res += fmt.Sprintf("    {\n       NutrientID: %s\n       Name: %s\n       Unit: %s\n       Value: %s\n    }\n",
				nutrient.NutrientID, nutrient.Name, nutrient.Unit, nutrient.Value)
		}
		res += fmt.Sprintf("  ]\n}\n")
	}
	return res
}

func getAPI(req *http.Request, client http.Client) []byte {
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request to API: %s\n", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body: %s\n", err)
	}

	return body
}

func getFoodsList(client http.Client, max string, apiKey string) interface{} {
	req, err := http.NewRequest("GET", apiURL+listEndpoint, nil)
	if err != nil {
		log.Fatalf("Error creating request: %s\n", err)
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("lt", "f")
	q.Add("max", max)
	q.Add("sort", "n")
	q.Add("api_key", apiKey)

	req.URL.RawQuery = q.Encode()

	body := getAPI(req, client)

	var fl FoodList

	erro := json.Unmarshal(body, &fl)
	if erro != nil {
		log.Fatalf("Error unmarshalling JSON: %s\n", erro)
	}
	return fl
}

func searchFood(client http.Client, food string, max string, apiKey string) interface{} {
	req, err := http.NewRequest("GET", apiURL+searchEndpoint, nil)
	if err != nil {
		log.Fatalf("Error creating request: %s\n", err)
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("sort", "r")
	q.Add("q", food)
	q.Add("max", max)
	q.Add("api_key", apiKey)

	req.URL.RawQuery = q.Encode()

	body := getAPI(req, client)

	var sl SearchList

	erro := json.Unmarshal(body, &sl)
	if erro != nil {
		log.Fatalf("Error unmarshalling JSON: %s\n", erro)
	}
	return sl
}

func getFoodByNdbno(client http.Client, ndbno string, apiKey string) FoodByNDBO {
	req, err := http.NewRequest("GET", apiURL+ndbnoEndpoint, nil)
	if err != nil {
		log.Fatalf("Error creating request: %s\n", err)
	}

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("type", "b")
	q.Add("ndbno", ndbno)
	q.Add("api_key", apiKey)

	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL)

	body := getAPI(req, client)

	var food FoodByNDBO

	erro := json.Unmarshal(body, &food)
	if erro != nil {
		log.Fatalf("Error unmarshalling JSON: %s\n", erro)
	}
	return food
}

func getToday() string {
	currentTime := time.Now().Local()
	return currentTime.Format("2006-01-02")
}

func removeFoodEntry(ndbno string) {
	dayToFoods := make(map[string]map[string]FoodToStore)
	date := getToday()
	if _, err := os.Stat("./calorietracker.json"); err == nil {
		raw, err := ioutil.ReadFile("./calorietracker.json")
		if err != nil {
			log.Fatalf("Could not read calorietracker.json %s\n", err)
		}
		json.Unmarshal(raw, &dayToFoods)
	} else {
		log.Fatalf("There is no calorietracker.json file to read %s\n", err)
	}
	if val, ok := dayToFoods[date]; ok {
		delete(val, ndbno)
		fmt.Println("Erased food record")
		b, err := json.MarshalIndent(dayToFoods, "", "\t")
		if err != nil {
			log.Fatalf("Error Marshalling struct into json %s\n", err)
		}

		err = ioutil.WriteFile("calorietracker.json", b, 0644)
		if err != nil {
			log.Fatalf("Error writing json to file %s\n", err)
		}
	} else {
		log.Fatalf("There is no record for the requested day")
	}
}

func showDayData(date string) {
	dayToFoods := make(map[string]map[string]FoodToStore)
	if _, err := os.Stat("./calorietracker.json"); err == nil {
		raw, err := ioutil.ReadFile("./calorietracker.json")
		if err != nil {
			log.Fatalf("Could not read calorietracker.json %s\n", err)
		}
		json.Unmarshal(raw, &dayToFoods)
	} else {
		log.Fatalf("There is no calorietracker.json file to read %s\n", err)
	}
	if val, ok := dayToFoods[date]; ok {
		b, err := json.MarshalIndent(val, "", "\t")
		if err != nil {
			log.Fatalf("Error Marshalling struct into json %s\n", err)
		}

		fmt.Printf("%+v\n", string(b[:]))
	} else {
		log.Fatalf("There is no record for the requested day")
	}
}

func addFood(client http.Client, ndbno string, apiKey string, qtd float64) {
	ndbnoResult := getFoodByNdbno(client, ndbno, apiKey)
	ndbnoToFood := make(map[string]FoodToStore)

	var f FoodToStore

	for _, food := range ndbnoResult.Foods {
		f.Name = food.Food.Desc.Name
		for _, nutrient := range food.Food.Nutrients {
			switch nutrient.NutrientID {
			case "208":
				i, err := strconv.ParseFloat(nutrient.Value, 16)
				if err != nil {
					log.Fatalf("Error converting string to int %s\n", err)
				}
				f.Energy = i
			case "203":
				i, err := strconv.ParseFloat(nutrient.Value, 16)
				if err != nil {
					log.Fatalf("Error converting string to int %s\n", err)
				}
				f.Protein = i
			case "204":
				i, err := strconv.ParseFloat(nutrient.Value, 16)
				if err != nil {
					log.Fatalf("Error converting string to int %s\n", err)
				}
				f.Fat = i
			case "205":
				i, err := strconv.ParseFloat(nutrient.Value, 16)
				if err != nil {
					log.Fatalf("Error converting string to int %s\n", err)
				}
				f.Carbohydrate = i
			case "291":
				i, err := strconv.ParseFloat(nutrient.Value, 16)
				if err != nil {
					log.Fatalf("Error converting string to int %s\n", err)
				}
				f.Fiber = i
			}
		}
	}

	dayToFoods := make(map[string]map[string]FoodToStore)

	if _, err := os.Stat("./calorietracker.json"); err == nil {
		raw, err := ioutil.ReadFile("./calorietracker.json")
		if err != nil {
			log.Fatalf("Could not read calorietracker.json %s\n", err)
		}
		json.Unmarshal(raw, &dayToFoods)
	}

	today := getToday()
	fmt.Println(today)

	if _, ok := dayToFoods[today]; !ok {
		// there is not previous data in that day
		// just point the key[date] to FoodByNdbno struct
		f.Qtd = qtd
		ndbnoToFood[ndbno] = f
		dayToFoods[today] = ndbnoToFood
	} else {
		// there is previous data for the day
		// verify if it already has the specific food entry
		if _, ok2 := dayToFoods[today][ndbno]; !ok2 {
			// if it doesnt't have, create one map and point dayToFoods[day] to it
			f.Qtd = qtd
			ndbnoToFood[ndbno] = f
			dayToFoods[today] = ndbnoToFood
		} else {
			// if it has, just add to the Qtd
			temp := dayToFoods[today][ndbno]
			temp.Qtd = temp.Qtd + qtd
			dayToFoods[today][ndbno] = temp
		}
	}

	b, err := json.MarshalIndent(dayToFoods, "", "\t")
	if err != nil {
		log.Fatalf("Error Marshalling struct into json %s\n", err)
	}

	err = ioutil.WriteFile("calorietracker.json", b, 0644)
	if err != nil {
		log.Fatalf("Error writing json to file %s\n", err)
	}

	return
}

func main() {
	max := flag.String("max", "50",
		"Maximum number of items to return\n")
	action := flag.String("action",
		"list", "Action (list, search, get_details, add, remove, show)")
	food := flag.String("food", "",
		"Food to search for or to add/remove from your daily track (name, ndbno).\n"+
			"Will be ignored if (show), (list) or (get_details) are selected on -action")
	day := flag.String("day", "",
		"Food to search for or to add/remove from your daily track (name, ndbno).\n"+
			"Format: yyyy-mm-dd\n"+
			"Will be ignored if an action different of (show) is selected on -action")
	ndbno := flag.String("ndbno", "",
		"Food Nbno to use on (get_details) action.\n"+
			"Will be ignored if an action different of (get_details) or (add) is selected on -action")
	qtd := flag.Float64("qtd", 100,
		"Weight of food consumed in grams (g).\n"+
			"Will be ignored if an action different of (add) is selected on -action")

	flag.Parse()

	apiKey := os.Getenv("USDA_API_KEY")
	client := http.Client{}

	switch *action {
	case "list":
		foodList := getFoodsList(client, *max, apiKey)
		fmt.Printf("%+v", foodList)
	case "search":
		searchList := searchFood(client, *food, *max, apiKey)
		fmt.Printf("%+v", searchList)
	case "get_details":
		ndbnoResult := getFoodByNdbno(client, *ndbno, apiKey)
		fmt.Printf("%+v", ndbnoResult)
	case "add":
		addFood(client, *ndbno, apiKey, *qtd)
	case "show":
		showDayData(*day)
	case "remove":
		removeFoodEntry(*ndbno)
	default:
		fmt.Println("Invalid action.\nValid actions are: (list, search, add, remove, show)")
		os.Exit(1)
	}

}
