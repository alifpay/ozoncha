package main

import (
	"alif/ozoncha/db"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zserge/lorca"
)

var ui lorca.UI

func main() {
	var (
		url = flag.String("url", "", "path of the config file")
		err error
	)
	flag.Parse()
	if len(*url) == 0 {
		return
	}

	err = db.Connect("gdb")
	defer db.Close()
	if err != nil {
		log.Fatalln("db error: ", err)
	}

	ui, err = lorca.New("", "", 1680, 800, "--log-level=3")
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()
	//grabLink(*url)
	grabTest(*url)
	/* go func() {
		time.Sleep(time.Second * 30)
		val := ui.Eval(`document.getElementById("PageCenter").innerHTML`)
		fmt.Print(val.String())
	}() */

	// Wait until UI window is closed
	<-ui.Done()
}

func grabLink(url string) {
	err := ui.Load(url)
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		time.Sleep(time.Minute * 5)
		val := ui.Eval(`document.getElementsByClassName("content")[0].innerHTML`)
		str := val.String()
		end := true
		for end {
			inx := strings.Index(str, `href="`)
			if inx == -1 {
				end = false
				break
			}
			inx += 6
			str = str[inx:]
			ix := strings.Index(str, `"`)
			//fmt.Println()
			v := str[:ix]
			if ln := len(v); ln > 10 && ln < 40 {
				_, err = db.Get("url" + v)
				if err != nil {
					db.Set([]byte("url"+v), []byte(v))
				}
			}
			str = str[ix:]
		}

	}()
}

func grabContent() {
	urls, err := db.GetAllBy("url")
	if err != nil {
		fmt.Println(err)
		return
	}
	records := [][]string{
		{"id", "title", "img", "price", "details", "Item Properties"},
	}
	goOn := make(chan bool)

	for _, url := range urls {
		url = "https://www.ozon.ru" + url
		idz := getIdz(url)
		err := ui.Load(url)
		if err != nil {
			fmt.Println(err)
		}

		go func() {
			time.Sleep(time.Second * 15)
			img := ui.Eval(`document.getElementsByClassName("eMicroGallery_fullImage")[0].src`)
			//fmt.Println(img.String())
			title := ui.Eval(`document.getElementsByClassName("bItemName")[0].innerText`)
			//fmt.Println(title.String())
			//eOzonPrice_main
			price := ui.Eval(`document.getElementsByClassName("eOzonPrice_main")[0].innerText`)
			//fmt.Println(price.String())

			//eItemDescription_text
			eItemDesc := ui.Eval(`document.getElementsByClassName("eItemDescription_text")[0].innerText`)
			//fmt.Println(eItemDesc.String())

			ui.Eval(`document.getElementsByClassName("eItemProperties_all mRoll jsShowAll")[0].click()`)

			//bItemProperties
			bItemProperties := ui.Eval(`document.getElementsByClassName("bItemProperties")[0].innerText`)
			//fmt.Println(bItemProperties.String())

			resp, err := http.Get(img.String())
			if err != nil {
				fmt.Println(err)
			}
			defer resp.Body.Close()

			// Create the file
			out, err := os.Create("./imgs/" + idz + ".jpg")
			if err != nil {
				fmt.Println(err)
			}
			defer out.Close()

			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			fmt.Println(err)
			//inx := strings.Index(str, `href="`)
			records = append(records, []string{idz, title.String(), idz + ".jpg", price.String(), strings.Replace(eItemDesc.String(), `"`, "", -1), strings.Replace(bItemProperties.String(), `"`, "", -1)})
			goOn <- true
		}()

		<-goOn
	}

	fileCsv, err := os.Create("./export.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer fileCsv.Close()

	w := csv.NewWriter(fileCsv)
	w.WriteAll(records) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}

func grabTest(url string) {

	inx := strings.Index(url, "id/")
	inx += 3
	idz := url[inx:]
	inx = len(idz) - 1
	idz = idz[:inx]
	fmt.Println(idz)
	err := ui.Load(url)
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		time.Sleep(time.Second * 15)
		//
		img := ui.Eval(`document.getElementsByClassName("eMicroGallery_fullImage")[0].src`)
		fmt.Println(img.String())
		title := ui.Eval(`document.getElementsByClassName("bItemName")[0].innerText`)
		fmt.Println(title.String())
		//eOzonPrice_main
		price := ui.Eval(`document.getElementsByClassName("eOzonPrice_main")[0].innerText`)
		fmt.Println(price.String())

		//eItemDescription_text
		eItemDesc := ui.Eval(`document.getElementsByClassName("eItemDescription_text")[0].innerText`)
		fmt.Println(eItemDesc.String())

		//eItemProperties_all mRoll jsShowAll
		ui.Eval(`document.getElementsByClassName("eItemProperties_all mRoll jsShowAll")[0].click()`)
		//fmt.Println(eItemDesc.String())
		//bItemProperties
		bItemProperties := ui.Eval(`document.getElementsByClassName("bItemProperties")[0].innerText`)
		fmt.Println(bItemProperties.String())

		resp, err := http.Get(img.String())
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create("./imgs/" + idz + ".jpg")
		if err != nil {
			fmt.Println(err)
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		fmt.Println(err)

		records := [][]string{
			{"id", "title", "img", "price", "details", "Item Properties"},
			{idz, title.String(), idz + ".jpg", price.String(), strings.Replace(eItemDesc.String(), `"`, "", -1), strings.Replace(bItemProperties.String(), `"`, "", -1)},
		}

		fileCsv, err := os.Create("./export.csv")
		if err != nil {
			fmt.Println(err)
		}
		defer fileCsv.Close()

		w := csv.NewWriter(fileCsv)
		w.WriteAll(records) // calls Flush internally

		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	}()
}

func getIdz(url string) string {
	inx := strings.Index(url, "id/")
	inx += 3
	idz := url[inx:]
	inx = len(idz) - 1
	idz = idz[:inx]
	return idz
}
