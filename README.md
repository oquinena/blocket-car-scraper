# blocket-car-scraper

CLI application written i Golang to extract time-series data from Swedish second-hand market blocket.se for selected car models. 

## Usage

Use with a runner (jenkins?) or as a cronjob once a day to generate either csv-files or json-data to stdout. 
eg. `./blocket-car-scraper -brand Volvo -model XC90 -output` to generate json-data containing ad subject, price, mileage etc. Omit the `-output` parameter to generate a csv file under ./data/Brand/Model/brand_model_yyyy-mm-dd.csv. 

## Helptext
```
This utility retrieves car ads from www.blocket.se and outputs to either stdout och csvfile.

  -brand string
    	Brand of the car you want to scrape. Have to be combined with either "-model" and/or "-list". Whitespaces have to be escaped, brandnames are case sensetive.
  -list
    	List available brands / models at blocket.se
  -model string
    	Model of the car brand you are looking for. Have to be combined with either "-brand" or "-list". Whitespaces have to be escaped, modelnames are case sensetive.
  -output
    	Log output as JSON to stdout
```