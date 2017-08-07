package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	searchURL = "https://www.avito.ru/sochi/kvartiry/prodam?pmax=4500000&pmin=2000000&district=209&f=549_5696-5697-5698.59_13986b.496_0b5124"
	streets   = []string{"победы", "калараш", "павлова", "партизанская", "лазарева", "коммунальников", "малышева", "тормахова", "местоположение"}
	pages     []string
	items     []itemType
	links     map[string]bool
)

type itemType struct {
	ID        string
	Title     string
	Link      string
	Address   string
	Price     int
	Comission string
}

func main() {
	var err error
	links = make(map[string]bool)
	err = parseURL(searchURL)
	if err != nil {
		log.Fatalln(err)
	}

	for needParse() {
		for _, p := range pages {
			if !links[p] {
				time.Sleep(time.Second)
				err = parseURL(p)
			}
		}
	}

	log.Println(len(items))
	log.Println(len(pages))
}

func needParse() bool {
	for _, p := range pages {
		if !links[p] {
			return true
		}
	}
	return false
}

func parseURL(urlstr string) error {
	log.Println(urlstr)
	client := new(http.Client)
	req, _ := http.NewRequest("GET", urlstr, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	links[urlstr] = true
	// doc, err := goquery.NewDocument(searchURL)
	if err != nil {
		log.Println("goquery.NewDocument", err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println(len(body))

	doc.Find(".catalog-list .item").Each(func(i int, s *goquery.Selection) {
		log.Println(i)
		var item itemType
		id, existID := s.Attr("id")
		item.ID = id
		if existID {
			a := s.Find("h3.title a")
			item.Title = strings.TrimSpace(a.Text())
			href, existtHREF := a.Attr("href")
			if existtHREF {
				item.Link = "https://www.avito.ru" + href
			}
			html, _ := s.Find(".about").Html()
			match := regexp.MustCompile(`.*<(div|a)`).FindString(html)

			item.Price, _ = strconv.Atoi(regexp.MustCompile(`[^\d]`).ReplaceAllString(match, ""))
			item.Address = strings.TrimSpace(s.Find(".address").Text())

			items = append(items, item)
			log.Println(len(items))
		} else {
			err = errors.New("Can not get attribute 'id'")
			log.Println(err)
		}
	})
	doc.Find("a.pagination-page").Each(func(i int, s *goquery.Selection) {
		href, existtHREF := s.Attr("href")
		if existtHREF {
			pages = append(pages, "https://www.avito.ru"+href)
		}
	})
	return err
}
