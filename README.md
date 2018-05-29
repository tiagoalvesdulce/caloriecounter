## Caloriecounter

Project developed to learn how to build a CLI using Google's Golang.

### Installing

`go get github.com/tiagoalvesdulce/caloriecounter`

### Adding an API key

Go to [USDA's website](https://ndb.nal.usda.gov/ndb/doc/index) and signup to get an API key.
Add the API key to your environment variables calling it `USDA_API_KEY`

### Usage

```
Usage of caloriecounter:
  -action string
    	Action (list, search, get_details, add, remove, show) (default "list")
  -day string
    	Food to search for or to add/remove from your daily track (name, ndbno).
    	Format: yyyy-mm-dd
    	Will be ignored if an action different of (show) is selected on -action
  -food string
    	Food to search for or to add/remove from your daily track (name, ndbno).
    	Will be ignored if (show), (list) or (get_details) are selected on -action
  -max string
    	Maximum number of items to return
    	 (default "50")
  -ndbno string
    	Food Nbno to use on (get_details) action.
    	Will be ignored if an action different of (get_details) or (add) is selected on -action
  -qtd float
    	Weight of food consumed in grams (g).
    	Will be ignored if an action different of (add) is selected on -action (default 100)
```

### Development

1. Fork the repo
2. `cd $GOPATH/src/github.com && git clone https://github.com/<your_github_name>/caloriecounter`

### To do

* Add food to daily calorie track
* Remove food from today's track
* Show calories of specific day
* Improve CLI interface
