package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func getFoodByNdbno(client http.Client, ndbno string, apiKey string) interface{} {
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

func main() {
	max := flag.String("max", "50",
		"Maximum number of items to return\n")
	action := flag.String("action",
		"list", "Action (list, search, get_details, add, remove, show)")
	food := flag.String("food", "",
		"Food to search for or to add/remove from your daily track (name, ndbno).\n"+
			"Will be ignored if (show), (list) or (get_details) are selected on -action")
	ndbno := flag.String("ndbno", "",
		"Food Nbno to use on (get_details) action.\n"+
			"Will be ignored if an action different of (get_details) is selected on -action")

	flag.Parse()

	apiKey := os.Getenv("USDA_API_KEY")
	client := http.Client{}

	fmt.Println(*ndbno)

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
	case "add", "remove", "show":
		fmt.Printf("Not implemented yet :(\nWant to contribute?\nGo to: https://github.com/tiagoalvesdulce/caloriecounter\n\n")
	default:
		fmt.Println("Invalid action.\nValid actions are: (list, search, add, remove, show)")
		os.Exit(1)
	}

}
