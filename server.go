package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

type PDFResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Error    string `json:"error"`
	Filename string `json:"filename"`
}

type PDFrequest struct {
	Html     string `json:"html"`
	Url      string `json:"url"`
	Title    string `json:"title"`
	Filename string `json:"filename"`
}

type RequestPayload struct {
	Html  string `json:"html" xml:"html" form:"html"`
	Url   string `json:"url" xml:"url" form:"url"`
	Title string `json:"title" xml:"title" form:"title"`
}

type ResponsePayload struct {
	Status     string `json:"status" xml:"status"`
	StatusCode string `json:"status_code" xml:"status_code"`
	Message    string `json:"message" xml:"message"`
	Errors     string `json:"errors" xml:"errors"`
}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	// Load the .env file in the current directory
	// load .env file
	godotenv.Load(".env")

	wkhtmlpath := os.Getenv("WKHTMLTOPDF_PATH")
	if len(wkhtmlpath) == 0 {
		fmt.Println("WKHTMLTOPDF_PATH is not set")
	}

	//Check if sepcific Prot required to start the server
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = renderer
	e.Static("/static", "public")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"name": "Rick James",
		})
	})

	e.POST("/html-to-pdf", func(c echo.Context) error {
		body := new(RequestPayload)
		if err := c.Bind(body); err != nil {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "Cannot parse input data",
				Errors:     err.Error(),
			})
		}

		// Assert if HTML is not empty
		if body.Html == "" {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "HTML input is required",
				Errors:     "Raw HTML input is required to generate the PDF",
			})
		}
		// Assert if Title is not empty
		if body.Title == "" {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "Title is required",
				Errors:     "A Title is required",
			})
		}
		//Obtain Timestamp for the current filename
		timestamp := time.Now().Unix()
		timestampstring := strconv.Itoa(int(timestamp))

		//grab the html inpt and create a static file in teh public directory
		filename := timestampstring + "_" + strings.Replace(body.Title, " ", "_", -1)
		err := os.WriteFile("./public/"+filename+".html", []byte(body.Html), 0644)
		if err != nil {

			c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "Error creating HTML file",
				Errors:     err.Error(),
			})
		}

		//Set the Url to the created file that is served statically
		body.Url = "http://localhost:" + port + "/static/" + filename + ".html"

		//Lets generate the PDF
		response := GeneratePDFDocument(PDFrequest{
			Url:      body.Url,
			Title:    body.Title,
			Html:     body.Html,
			Filename: filename,
		})

		return c.JSON(http.StatusCreated, response)
	})

	e.POST("/url-to-pdf", func(c echo.Context) error {
		body := new(RequestPayload)
		if err := c.Bind(body); err != nil {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "Cannot parse input data",
				Errors:     err.Error(),
			})
		}

		// Assert if url is not empty
		if body.Url == "" {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "URL input is required",
				Errors:     "A URL to generate the PDF is required",
			})
		}
		// Assert if Title is not empty
		if body.Title == "" {
			return c.JSON(http.StatusBadRequest, &ResponsePayload{
				Status:     "error",
				StatusCode: "400",
				Message:    "Title is required",
				Errors:     "A Title is required",
			})
		}

		//Obtain Timestamp for the current filename
		timestamp := time.Now().Unix()
		timestampstring := strconv.Itoa(int(timestamp))
		//grab the html inpt and create a static file in teh public directory
		filename := timestampstring + "_" + strings.Replace(body.Title, " ", "_", -1)
		//Lets generate the PDF
		response := GeneratePDFDocument(PDFrequest{
			Url:      body.Url,
			Title:    body.Title,
			Html:     body.Html,
			Filename: filename,
		})

		return c.JSON(http.StatusCreated, response)
	})

	e.Logger.Fatal(e.Start(":" + port))
}

/*
Function to Generate the PDF WIth the WKHTMLTOPDF Webkit engine
*/
func GeneratePDFDocument(req PDFrequest) (res PDFResponse) {

	fmt.Println("WKHTMLTOPDF_PATH:", os.Getenv("WKHTMLTOPDF_PATH"))
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Println("Error")
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}

	// Set global options
	pdfg.Dpi.Set(72)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)

	// Create a new input page from an URL
	page := wkhtmltopdf.NewPage(req.Url)

	// Set options for this page
	page.FooterRight.Set("[page]")
	page.FooterFontSize.Set(10)
	page.Zoom.Set(0.95)

	// Add to document
	pdfg.AddPage(page)

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}

	//CReate filename and escape special characters from the title
	filename := req.Filename + ".pdf"
	// Write buffer contents to file on disk
	err = pdfg.WriteFile("./public/" + filename)
	if err != nil {
		res := PDFResponse{
			Status:  "error",
			Message: "Error creating PDF",
			Error:   err.Error(),
		}
		return res
	}
	//Return the response
	return PDFResponse{
		Status:   "success",
		Message:  "PDF Created Successfully",
		Filename: "static/" + filename,
		Error:    "",
	}

}
