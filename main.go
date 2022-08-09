package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Struct for available car manufacturers and models
type Car struct {
	CategoryCounters []struct {
		SearchParameters string `json:"search_parameters"`
		Label            string `json:"label"`
		APIQuery         string `json:"api_query"`
		AdCounter        int    `json:"ad_counter"`
	} `json:"category_counters"`
}

// Struct for car ads on blocket.se
type Ad struct {
	Data []struct {
		AdID       string `json:"ad_id"`
		AdStatus   string `json:"ad_status"`
		Attributes []struct {
			Header string   `json:"header"`
			ID     string   `json:"id"`
			Items  []string `json:"items"`
		} `json:"attributes"`
		LicensePlate string `json:"license_plate"`
		ListID       string `json:"list_id"`
		Location     []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			QueryKey string `json:"query_key"`
		} `json:"location"`
		ParameterGroups []struct {
			Label      string `json:"label"`
			Parameters []struct {
				ID    string `json:"id"`
				Label string `json:"label"`
				Value string `json:"value"`
			} `json:"parameters"`
			Type string `json:"type"`
		} `json:"parameter_groups"`
		Price struct {
			Label  string `json:"label"`
			Suffix string `json:"suffix"`
			Value  int    `json:"value"`
		} `json:"price"`
		ShareURL string `json:"share_url"`
		Subject  string `json:"subject"`
	} `json:"data"`
}

func main() {
	// CLI switches & usage
	carBrand := flag.String("brand", "", "Brand of the car you want to scrape. Have to be combined with either \"-model\" and/or \"-list\". Whitespaces have to be escaped, brandnames are case sensetive.")
	carModel := flag.String("model", "", "Model of the car brand you are looking for. Have to be combined with either \"-brand\" or \"-list\". Whitespaces have to be escaped, modelnames are case sensetive.")
	list := flag.Bool("list", false, "List available brands / models at blocket.se")
	outdir := flag.String("outdir", "", "Set output directory for csv files")
	logStdOut := flag.Bool("output", false, "Log output as JSON to stdout")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "This utility retrieves car ads from www.blocket.se and outputs to either stdout och csvfile.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Print usage if no flags are set (same as -h)
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	// Make sure -model or -list is set when specifying brand
	if *carBrand != "" && *carModel == "" && !*list {
		flag.Usage()
		os.Exit(0)
	} else if *carBrand == "" && *carModel != "" && !*list {
		flag.Usage()
		os.Exit(0)
	}

	token, err := getBearerToken()
	if err != nil {
		log.Printf("Error requesting BearerToken\n%s", err)
	}
	brand, err := getCarBrand(token)
	if err != nil {
		log.Printf("Error retrieving car brand\n%s", err)
	}

	// Match url with string from argument -brand
	var selectedBrand string
	var selectedBrandName string
	for _, v := range brand.CategoryCounters {
		if v.Label == *carBrand {
			selectedBrand = v.SearchParameters
			selectedBrandName = v.Label
		}
	}
	model, err := getBrandModel(token, selectedBrand)
	if err != nil {
		log.Printf("%s", err)
	}

	var selectedModel string
	var selectedModelName string
	for _, v := range model.CategoryCounters {
		if v.Label == *carModel {
			selectedModel = v.SearchParameters
			selectedModelName = v.Label
		}
	}

	// List available brands or models
	if *list {
		if *carBrand != "" {
			for _, v := range model.CategoryCounters {
				fmt.Println(v.Label)
			}
		} else {
			for _, v := range brand.CategoryCounters {
				fmt.Println(v.Label)
			}
		}
	}

	ads := getListedAds(token, selectedModel)

	if *logStdOut && !*list && *outdir == "" {
		fmt.Println(PrettyPrint(ads))
	} else if !*list {
		if *outdir == "" {
			*outdir, _ = os.Getwd()
		}
		err = outputToCSV(selectedBrandName, selectedModelName, ads, *outdir)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
		}
	}
}

// Make request to blocket.se and extract bearerToken to use with api.blocket.se
func getBearerToken() (string, error) {
	url := "https://blocket.se/"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	// We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	// Convert the body to type string
	sb := string(body)
	// Split string at first instance of "bearerToken"
	tmp := strings.Split(sb, "bearerToken")
	// Split second slice by "-character
	tmp2 := strings.Split(tmp[1], "\"")
	// return stripped bearerToken
	return tmp2[2], nil
}

// Retrieve all car manufacturers listed on blocket.se
func getCarBrand(token string) (Car, error) {
	url := "https://api.blocket.se/classifieds/v1/ad_counters?cg=1020&include=all"
	auth := "Bearer " + token

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Car{}, err
	}
	req.Header.Set("authorization", auth)
	response, err := client.Do(req)
	if err != nil {
		return Car{}, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body) // response body is []byte
	if err != nil {
		return Car{}, err
	}

	var brand Car
	if err := json.Unmarshal(body, &brand); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return brand, nil
}

// Retrieve all models for given manufacturer
func getBrandModel(token, brandSearchParams string) (Car, error) {
	url := "https://api.blocket.se/classifieds/v1/ad_counters?" + brandSearchParams + "&include=all"
	auth := "Bearer " + token

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Car{}, err
	}
	req.Header.Set("authorization", auth)
	response, err := client.Do(req)
	if err != nil {
		return Car{}, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body) // response body is []byte
	if err != nil {
		return Car{}, err
	}

	var model Car
	if err := json.Unmarshal(body, &model); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	return model, nil
}

// Retrieve all ads for given manufacturer / model
func getListedAds(token, modelSearchParams string) Ad {
	url := "https://api.blocket.se/search_bff/v1/content?" + modelSearchParams + "&include=all"
	auth := "Bearer " + token

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
	}
	req.Header.Set("authorization", auth)
	response, err := client.Do(req)
	if err != nil {
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body) // response body is []byte
	if err != nil {
	}

	var ad Ad
	if err := json.Unmarshal(body, &ad); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return ad

}

// Output data to CSV-file
func outputToCSV(brandName string, modelName string, ads Ad, outdir string) error {
	todaysDate := time.Now()
	filename := fmt.Sprintf("%s_%s_%d-%02d-%02d.csv", brandName, modelName, todaysDate.Year(), todaysDate.Month(), todaysDate.Day())

	// Create directory data/brand/model if not exists
	if _, err := os.Stat(filepath.Join(outdir, "data", brandName, modelName)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(outdir, "data", brandName, modelName), 0755)
	}

	csvFile, err := os.Create(filepath.Join(outdir, "data", brandName, modelName, filename))
	if err != nil {
		return err
	}

	csvwriter := csv.NewWriter(csvFile)
	columnnames := []string{"Subject", "Price", "Mileage", "Year", "Municipality", "Area", "URL"}
	err = csvwriter.Write(columnnames)
	if err != nil {
		fmt.Println("ERROR Writing columnnames")
	}

	for _, v := range ads.Data {
		var line []string
		line = append(line, v.Subject)
		line = append(line, fmt.Sprint(v.Price.Value))
		line = append(line, v.ParameterGroups[0].Parameters[2].Value)
		line = append(line, v.ParameterGroups[0].Parameters[3].Value)
		line = append(line, v.Location[0].Name)
		if len(v.Location) > 1 {
			line = append(line, v.Location[1].Name)
		} else {
			line = append(line, "N/A")
		}
		line = append(line, v.ShareURL)
		err := csvwriter.Write(line)
		if err != nil {
			fmt.Println(err)
		}
	}
	csvwriter.Flush()
	csvFile.Close()
	return nil
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
